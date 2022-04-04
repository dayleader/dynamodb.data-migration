# DynamoDB data migration tool

Data migration tool for Amazon DynamoDB. Migration files are described as json files and can contain both schema and data migrations.

## Usage

| Flag       |   Default value     | Description |
|------------|---------------------|-------------|
| `migrations` | `/migrations` | Directory where the migration files are located (should not be hierarchical) |
| `x-migrations-table` | `x_migrations` | Name of the migrations table |

Execute the binary:

    ./migrations --migrations=example/migrations

Exit codes:

    0 - Migration succeeded
    1 - Migration failed

AWS variables:

 * AWS_REGION - aws region
 * AWS_ACCESS_KEY_ID  - aws key
 * AWS_SECRET_ACCESS_KEY - aws secret    

## Using the Docker Image

The docker image has one volume: **/migrations** which is the directory containing your json files.

    docker run --rm -v $(pwd)/examples/migrations/dev:/migrations -e MIGRATIONS_DIR=${MIGRATIONS_DIR} -e MIGRATIONS_TABLE_NAME=${MIGRATIONS_TABLE_NAME} -e AWS_REGION=${AWS_REGION} -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}

> Note: **/migrations** volume should not be hierarchical.

Environment variables:
 
 * MIGRATIONS_DIR - directory where the migration files are located
 * MIGRATIONS_TABLE_NAME - name of the migrations table

AWS variables:

 * AWS_REGION - aws region
 * AWS_ACCESS_KEY_ID  - aws key
 * AWS_SECRET_ACCESS_KEY - aws secret

## Migration naming convention

All changes to the database are called migrations. Migrations have a version and a description. The version must be unique inside the environment. 
Migrations will be applied in ascending order by version and exactly once.

The ordering and direction of the migration files is determined by the filenames used for them. The tool expects the filenames of migrations to have the format:

    {version}_{title}.{extension}

> Version - semantic versioning, major.minor.path (e.g. 1.156.0, 1.156.1 ...)

The title of each migration is unused, and is only for readability. Similarly, the extension of the migration files is not checked by the library, and should be an appropriate format for the database in use (.json).

Examples of valid migration file naming:
    
    1.156.0_create_users.json // will be applied first and only once.
    1.156.1_0_create_roles.json // will be applied after the version 1.156.0 and only once.

## JSON statement format

Example of valid statement:

    [
        {
            "table_name": "users",
            "schema": [
                {
                    "attribute_definitions": [
                        {
                            "name": "id",
                            "type": "S"
                        }
                    ],
                    "key_schema": [
                        {
                            "name": "id",
                            "type": "HASH"
                        }
                    ]
                }
            ],
            "data": [
                {
                    "id": "1",
                    "username": "test1",
                    "email": "test1@email.com"
                },
                {
                    "id": "2",
                    "username": "test2",
                    "email": "test2@email.com"
                },
                {
                    "id": "3",
                    "username": "test3",
                    "email": "test3@email.com"
                }
            ]
        }
    ]

## Migration execution

Each migration file will be executed in one transaction and only once inside the environment.
After the migration completed, in the database will be created a migration record for each migration file.

The schema will be migrated first, and then all data migrations will be done in a single transaction. Please read the wirte transaction limitations: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/transaction-apis.html#transaction-apis-txwriteitems

Example of migration record:

    {
      "version": "1.156.0",
      "name": "1.156.0_create_users.json"
      "start_time": 1626681490, // unix
      "execution_time": 2, // seconds
    }
