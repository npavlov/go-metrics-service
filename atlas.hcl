variable "path" {
  type        = string
  description = "A path to the template directory"
}

data "template_dir" "schema" {
  path = var.path
  vars = {
    key = "value"
    // Pass the --env value as a template variable.
    env  = atlas.env
  }
}

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
    src = data.template_dir.schema.url
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