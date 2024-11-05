

data "external" "dot_env" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./cmd/envloader",
  ]
 }

locals {
  dot_env = jsondecode(data.external.dot_env)
}

env "dev" {
    src = "sql/schema.tmpl.sql"
    dev = local.dot_env.TEMP_DB
    migration {
        dir = "file://migrations?format=goose"
    }
    format {
        migrate {
            diff = "{{ sql . \"  \" }}"
        }
    }
}