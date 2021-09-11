
resource "aws_ecs_cluster" "ecs_cluster" {
  name  = "postponer-cluster"
}

/*
resource "aws_ecs_task_definition" "task_definition" {
  family                = "service"
  cpu = "256"
  memory = "512"
  execution_role_arn = "arn:aws:iam::224040641882:role/ecsTaskExecutionRole"
  task_role_arn = "arn:aws:iam::224040641882:role/ecsTaskExecutionRole"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  container_definitions = jsonencode([
    {
      name      = "postponer"
      image     = "224040641882.dkr.ecr.eu-central-1.amazonaws.com/postponer:0.0.1"
//      cpu       = 2
//      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 80
          hostPort      = 80
        }
      ]
      environment = [
        {
          name = "DB_DSN"
          value = aws_db_instance.postgres.endpoint
        }
      ]
    }
  ])
}

resource "aws_ecs_service" "postponer" {
  name            = "postponer"
  cluster         = aws_ecs_cluster.ecs_cluster.id
  task_definition = aws_ecs_task_definition.task_definition.arn
  desired_count   = 2
  launch_type = "FARGATE"
  network_configuration {
    subnets = [aws_subnet.pub_subnet.id, aws_subnet.pub_subnet-2.id]
  }
}
*/