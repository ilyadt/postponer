package main

import (
    "context"
    "database/sql"
    "net/http"
    "os"
    "os/signal"
    "postponer/core"
    "postponer/providers/sqlstorage"
    "postponer/providers/stdlogger"
    "postponer/providers/stdoutdispatcher"
    "sync"
    "syscall"
    "time"

    _ "github.com/lib/pq"
)

func main() {

    dispatcher := &stdoutdispatcher.StdoutDispatcher{}

    logger := &stdlogger.StdLogger{}

    dbDsn := os.Getenv("DB_DSN")
    if len(dbDsn) == 0 {
        panic("No env DB_DSN provided")
    }

    db, err := sql.Open("postgres", dbDsn)
    if err != nil {
        panic("Failed to connect to DB error: " + err.Error())
    }
    defer func() {
        if err := db.Close(); err != nil {
            logger.Errorf("Close db connection error: %s", err.Error())
        }
    }()

    // Ограничение сверху на кол-во коннектов в пуле
    db.SetMaxOpenConns(50)

    storage := sqlstorage.NewStorage(db, logger)

    backgroundService := core.NewBackgroundService(storage, dispatcher)

    var wg sync.WaitGroup

    // background service
    ctx, stopBackground := context.WithCancel(context.Background())

    wg.Add(1)
    go func() {
        defer wg.Done()
        backgroundService.Do(ctx)
    }()

    // http-server
    handler := core.NewAddHandler(dispatcher, storage, backgroundService)

    // TODO: go-lint

    // TODO: добавить логирование запросов, прометеус
    // Статус ответа по методу
    // время ответа
    // количество запросов rate

    // количество задач в очереди по queue
    // количество обработанных сообщений
    // количество ошибок dispatcher, storage

    // количество запросов в базу / длительность запроса
    // количество релоудов background

    // логирование запросов
    // логирование отправки в dispatcher

    // TODO: maxMessageLength, maxDelay

    serveMux := new(http.ServeMux)

    serveMux.HandleFunc("/add", handler.Request)

    httpServer := &http.Server{
        Addr:              "0.0.0.0:80",
        Handler:           serveMux,
        ReadHeaderTimeout: 30 * time.Second, // Mitigate Slowloris Attack
    }

    wg.Add(1)

    go func() {
        defer wg.Done()
        err := httpServer.ListenAndServe()
        if err != nil && err != http.ErrServerClosed {
            panic("Error Starting the HTTP Server : " + err.Error())
        }
    }()

    logger.Info("Started...")

    // Waiting for stop signal
    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    <-done

    // Stopping http server
    logger.Info("Stopping http server...")
    if err := httpServer.Shutdown(context.Background()); err != nil {
        logger.Info("Server Shutdown Failed: " + err.Error())
    }

    // Stopping background service
    logger.Info("Stopping background service...")
    stopBackground()

    // Wait till server and background service gracefully finishing
    wg.Wait()

    logger.Info("Bye-Bye")
}
