#!/bin/bash
while ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; do
  echo "Waiting for database to be ready..."
  sleep 1
done

echo "Database is ready, starting the app!"
/pvz
