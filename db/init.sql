CREATE TABLE postponer_queue (
      id VARCHAR(255) NOT NULL primary key,
      queue VARCHAR(255) NOT NULL,
      body  TEXT NOT NULL,
      fires_at BIGINT not null,
      created_at BIGINT not null
);

CREATE INDEX fires_at_idx ON postponer_queue(fires_at);
CREATE INDEX id_idx ON postponer_queue(id);
