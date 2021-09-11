
/*
resource "aws_db_subnet_group" "db_subnet_group" {
  subnet_ids  = [aws_subnet.pub_subnet.id, aws_subnet.pub_subnet-2.id]
}

resource "aws_db_instance" "postgres" {
  identifier                = "test"
  allocated_storage         = 20 # 2GiB of storage space
  backup_retention_period   = 2
  backup_window             = "01:00-01:30"
  maintenance_window        = "sun:03:00-sun:04:30"
  multi_az                  = false
  engine                    = "postgres"
  engine_version            = "12.5"
  instance_class            = "db.t3.micro"
//  name                      = "worker_db"
  username                  = "worker"
  password                  = "worker_so_worker_1$!"
  port                      = "5432"
  db_subnet_group_name      = aws_db_subnet_group.db_subnet_group.id
  vpc_security_group_ids    = [aws_security_group.rds_sg.id, aws_security_group.ecs_sg.id]
  skip_final_snapshot       = true
  final_snapshot_identifier = "worker-final"
  publicly_accessible       = true
}

output "postgres_resource_id" {
  value = aws_db_instance.postgres.id
}

output "postgres_endpoint" {
  value = aws_db_instance.postgres.endpoint
}
*/