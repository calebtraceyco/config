# config-yaml

---
See example usage [here](https://github.com/CalebTracey/myb-api/tree/main/cmd/svr)
## Install

    go get -u github.com/calebtracey/config-yaml

### Example yaml file config:

``` yaml
Env: "Dev"
Port: 6080
AppName: "example-app"
ComponentConfigs:
  Client:
    Timeout: 15
    IdleConnTimeout: 15
    MaxIdleConsPerHost: 50
    MaxConsPerHost: 0
    DisableCompression: 2
    InsecureSkipVerify: 1
Services:
  - Name: "TestAPI"
    URL: "https://api.test.net/v5"
    ApiKeyEnvironmentVariable: "API_KEY"
    PublicKeyEnvironmentVariable: "PUBLIC_KEY"
Databases:
  - Name: "PostgresDB"
    Database: "postgres"
    Server: "db.example.supabase.co:5432"
    Username: "postgres"
    Scheme: "postgres"
    AuthRequired: "true"
    AuthEnvironmentVariable: "DB_PASSWORD_DEV"
  ```

