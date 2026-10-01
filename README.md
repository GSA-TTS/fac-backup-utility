# go-backup

## while developing/testing

```
go run main.go
```

## to build

```
./build.sh
```

This command clones one DB to another by piping STDOUT from `pg_dump` into the STDIN of `psql`, with the correct connection/credential parameters for each command.

When running locally (assuming `ENV` is set to `LOCAL`) it will read a `config.json` from the directory `$HOME/.fac/config.json` (or, from `config.json` in the same folder as the application). This file should look like a `VCAP_SERVICES` variable that would be encountered in the Cloud Foundry/cloud.gov environment.

When run in the cloud.gov environment (where `ENV` is anything other than `LOCAL` or `TESTING`), it will look at `$VCAP_SERVICES`, look in the `aws-rds` key, and look up the DB credentials by the friendly name provided on the command line. By this, if your brokered DB is called `fac-db`, this will then populate the credentials (internally) with the brokered DB name, password, URI, etc. in order to correctly `pg_dump` from one and, using another set of credentials, stream the data into another.

This does *not* guarantee a perfect backup. It *does* do a rapid snapshot at a moment in time, without requiring the application to write any files to the local filesystem within a container. (On cloud.gov, this is limited to ~6GB, which makes dumping and loading DBs difficult.)

## Usage: bucket

```
gsa.gov.fac.cgov-util bucket --source-db <name> --destination-bucket <name>
```

Similar to above, but this pipes a `pg_dump` to `s3 copy`.

For now, this writes to the key `s3://<bucket>/backups/<name>-<db-name>.dump`

This wants to be improved.

The purpose here is to (again) dump a database to a storage location without touching the local (containerized) filesystem. It uses friendly names, again, to look up the credentials for both the RDS database and brokered S3 in order to stream a DB dump to S3. (In theory, S3 does multipart uploads, so you should end up with a single file, up to 5TB in size, for your dump.)

When running locally, this assumes `minio` is running as a stand-in for S3, and is specified as a `user-specified` service in the (local, bogus) VCAP_SERVICES config.

(An example `config.json` is in this repository, and a more complete file in `internal/vcap/vcap_test.go`).


## Assumptions

* The `ENV` var is set to `LOCAL` for local testing. i.e `export ENV="LOCAL"`
* You have two Postgres containers running, one at port 5432, and another at 5431.

You can change the local DB values in `config.yaml` to reflect your config.

In a remote environment, the variable `VCAP_SERVICES` is referenced to extract values.

## Minio on Windows
- Open powershell as administrator to download the tool.
- Move `C:\mc.exe` to the root of the project folder.
```
Invoke-WebRequest -Uri "https://dl.minio.io/client/mc/release/windows-amd64/mc.exe" -OutFile "C:\mc.exe"
```

## Adding a New Command
- We utilize cobra-cli as a basis for creating commands in [cmd](./cmd/). It will generate a template file for use, prepoulated with necessary information to start building a command.
- More information can be found at the [cobra-cli readme](https://github.com/spf13/cobra-cli/blob/main/README.md)
```
cobra-cli add <command_name>
```

## Common Command Usage

- Fetch the latest release when running on a cloud.gov environment:
```
# With Proxy:
curl -x $https_proxy -L "https://github.com/GSA-TTS/fac-backup-utility/releases/download/vX.Y.ZZ/gov.gsa.fac.cgov-util-vX.Y.ZZ-linux-amd64.tar.gz" -O

# Without Proxy
curl -L "https://github.com/GSA-TTS/fac-backup-utility/releases/download/vX.Y.ZZ/gov.gsa.fac.cgov-util-vX.Y.ZZ-linux-amd64.tar.gz" -O

tar -xvf gov.gsa.fac.cgov-util-vX.Y.ZZ-linux-amd64.tar.gz && rm gov.gsa.fac.cgov-util-vX.Y.ZZ-linux-amd64.tar.gz
```

- Install AWS CLI on cloud.gov instances:
    - It is advised to not run this on a local machine, due to where aws will be installed and could potentially add conflicts. Please install AWS CLI on your local environment using the official methods provided by AWS for your OS.
```
./gov.gsa.fac.cgov-util install_aws
```

- Use AWS S3 Sync to sync the contents of two s3 buckets:
```
./gov.gsa.fac.cgov-util s3_sync --source_s3 s3://<src-s3-name>/ --dest_s3 s3://<dest-s3-name>/
```

- Backup an Postgres instance to an s3 using psql .bin files:
```
./gov.gsa.fac.cgov-util s3_to_db --db <src-db-name> --s3path s3://<dest-s3-name>/path/to/store/
```

- Backup Postgres Tables to another Postgres instance:
    - This requires a secondary postgres in your docker compose, with the expected `5431:5432` ports, while the primary runs on `5432:5432`. These can be changed if desired. Port changes (if applicable) are only required for local testing & development.
    - When running on cloud.gov environements, the port and URI are contained in `$VCAP_SERVICES` for the `src_db` and `dest_db` and thus require no changes.
```
./gov.gsa.fac.cgov-util db_to_db --src_db <src-db-name> --dest_db <dest-db-name>
```

- Check db for missing tables, based on the [manifest](./cmd/assets/db_tables.txt)
```
./gov.gsa.fac.cgov-util check_db --db <db-name>
```

- Check row count for a table existing in the db, based on the [manifest](./cmd/assets/db_tables.txt)
```
./gov.gsa.fac.cgov-util row_count --db <db-name>
```
