# Cloud Migration Workflow Management
This is a subsystem of the Cloud-Barista platform that provides workflow management for cloud migration.

## Overview

* Create and management workflow through Airflow.
* Create workflow based on gusty.

## Development environment
* Tested operating systems (OSs):
    * Ubuntu 24.04, Ubuntu 22.04, Ubuntu 18.04
* Language:
    * Go: 1.23.0

## How to run

### 1. Build and run

Build and run the binary with Airflow server
```shell
make run
```

Or, you can run it within Docker by this command.
 ```shell
 make run_docker
 ```

If you want to stop the binary with Airflow server, run this command.
```shell
make stop
```

### 2. Write configuration file
 
- Configuration file name is 'cm-cicada.yaml'
- The configuration file must be placed in one of the following directories.
    - .cm-cicada/conf directory under user's home directory
    - 'conf' directory where running the binary
    - 'conf' directory where placed in the path of 'CMCICADA_ROOT' environment variable
- Configuration options
    - task_component
        - load_examples : Load task component examples if true.
        - examples_directory : Specify directory where task component examples are located. Must be set if 'load_examples' is true.
    - workflow_template
        - templates_directory : Specify directory where workflow templates are located.
    - airflow-server
        - address : Specify Airflow server's address ({IP or Domain}:{Port})
        - use_tls : Must be true if Airflow server uses HTTPS.
        - skip_tls_verify : Skip TLS/SSL certificate verification. Must be set if 'use_tls' is true.
        - init_retry : Retry count of initializing Airflow server connection used by cm-cicada.
        - timeout : HTTP timeout value as seconds.
        - username : Airflow login username.
        - password : Airflow login password.
        - connections : Pre-define Airflow connections (Set multiple connections)
          - id : ID of connection
          - type : Type of connection
          - description : Description of connection
          - host : Host address or URL of connection
          - port : Port number for use connection
          - schema : Connection schema
          - login : Username for use connection
          - password : Password for use connection
    - dag_directory_host : Specify DAG directory of the host. (Mounted DAG directory used by Airflow container.)
    - dag_directory_container : Specify DAG directory of Airflow container. (DAG directory inside the container.)
- listen
    - port : Listen port of the API.
- Configuration file example
  ```yaml
  cm-cicada:
    task_component:
        load_examples: true
        examples_directory: "./lib/airflow/example/task_component/"
    workflow_template:
        templates_directory: "./lib/airflow/example/workflow_template/"
    airflow-server:
        address: 127.0.0.1:8080
        use_tls: false
        # skip_tls_verify: true
        init_retry: 5
        timeout: 10
        username: "airflow"
        password: "airflow_pass"
        connections:
          - id: honeybee_api
            type: http
            description: HoneyBee API
            host: 127.0.0.1
            port: 8081
            schema: http
          - id: beetle_api
            type: http
            description: Beetle API
            host: 127.0.0.1
            port: 8056
            schema: http
            login: default
            password: default
          - id: tumblebug_api
            type: http
            description: TumbleBug API
            host: 127.0.0.1
            port: 1323
            schema: http
            login: default
            password: default
    dag_directory_host: "./_airflow/airflow-home/dags"
    dag_directory_container: "/usr/local/airflow/dags" # Use dag_directory_host for dag_directory_container, if this value is empty
    listen:
        port: 8083
  ```

### 3. Check workflow template

Check workflow template list.

```shell
curl -X 'GET' \
  'http://127.0.0.1:8083/cicada/workflow_template' \
  -H 'accept: application/json'
```

Get workflow template and copy the content.

```shell
curl -X 'GET' \
  'http://127.0.0.1:8083/cicada/workflow_template/81bbeb23-2c48-4536-9f01-55796e0fa394' \
  -H 'accept: application/json'
```

### 4. Create the workflow

Create the workflow by pasting copied workflow template content.
Modify all of 'sgId' params. (Create the source group from honeybee.)

Check the ID of the created workflow.

```shell
curl -X 'POST' \
  'http://127.0.0.1:8083/cicada/workflow' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "migrate_infra_workflow",
  "data": {
    "description": "Migrate Server",
    "task_groups": [
      {
        "name": "migrate_infra",
        "description": "Migrate Server",
        "tasks": [
          {
            "name": "infra_import",
            "task_component": "honeybee_task_import_infra",
            "request_body": "",
            "path_params": {
              "sgId": "ba695bbb-d673-4092-9821-c9cb05676228"
            },
            "dependencies": []
          },
          {
            "name": "infra_get",
            "task_component": "honeybee_task_get_infra_refined",
            "request_body": "",
            "path_params": {
              "sgId": "ba695bbb-d673-4092-9821-c9cb05676228"
            },
            "dependencies": [
              "infra_import"
            ]
          },
          {
            "name": "infra_recommend",
            "task_component": "beetle_task_recommend_infra",
            "request_body": "infra_get",
            "path_params": null,
            "dependencies": [
              "infra_get"
            ]
          },
          {
            "name": "infra_migration",
            "task_component": "beetle_task_infra_migration",
            "request_body": "infra_recommend",
            "path_params": null,
            "dependencies": [
              "infra_recommend"
            ]
          },
          {
            "name": "register_target_to_source_group",
            "task_component": "honeybee_register_target_info_to_source_group",
            "request_body": "infra_migration",
            "path_params": {
              "sgId": "ba695bbb-d673-4092-9821-c9cb05676228"
            },
            "dependencies": [
              "infra_migration"
            ]
          }
        ]
      }
    ]
  }
}'
```

### 5. Run the workflow

```shell
curl -X 'POST' \
  'http://127.0.0.1:8083/cicada/workflow/4420da6c-c50f-4d8b-bc2e-f02d8b557fad/run' \
  -H 'accept: application/json' \
  -d ''
```

## About Task Component
Each task in the workflow references a Task Component.

The Task Component part frequently leads to mistakes, and when the Request Body structure is complex or changes, it becomes difficult to keep up. Therefore, we have implemented a feature that automatically generates Task Components by reading JSON files.

As shown below, using JSONs containing name, description, api_connection_id, swagger_yaml_endpoint, and endpoint, the cicada reads the Swagger YAML and finds the endpoint to automatically construct task components.

The api_connection_id corresponds to one of the connection ids defined in https://github.com/cloud-barista/cm-cicada/blob/main/conf/cm-cicada.yaml

```json
{
  "name": "tumblebug_mci_dynamic",
  "description": "Create MCI Dynamically from common spec and image.",
  "api_connection_id": "tumblebug_api",
  "swagger_yaml_endpoint": "/tumblebug/api/doc.yaml",
  "endpoint": "/ns/{nsId}/mciDynamic"
}
```

Examples of these JSONs can be found at:
https://github.com/cloud-barista/cm-cicada/tree/main/lib/airflow/example/task_component

The Task Component automatically generated from the above JSON is as follows:

<details>
    <summary>Expand</summary>

```json
{
  "id": "796645fc-c594-4263-a9ff-243051d1f3a5",
  "name": "tumblebug_mci_dynamic",
  "description": "Create MCI Dynamically from common spec and image.",
  "data": {
    "options": {
      "api_connection_id": "tumblebug_api",
      "endpoint": "/tumblebug/ns/{nsId}/mciDynamic",
      "method": "POST",
      "request_body": "{\n    \"description\": \"Made in CB-TB\",\n    \"installMonAgent\": \"no\",\n    \"label\": {},\n    \"name\": \"mci01\",\n    \"systemLabel\": \"\",\n    \"vm\": [\n        {\n            \"commonImage\": \"ubuntu18.04\",\n            \"commonSpec\": \"aws+ap-northeast-2+t2.small\",\n            \"connectionName\": \"string\",\n            \"description\": \"Description\",\n            \"label\": {},\n            \"name\": \"g1-1\",\n            \"rootDiskSize\": \"default, 30, 42, ...\",\n            \"rootDiskType\": \"default, TYPE1, ...\",\n            \"subGroupSize\": \"3\",\n            \"vmUserPassword\": \"string\"\n        }\n    ]\n}"
    },
    "body_params": {
      "required": [
        "name",
        "vm"
      ],
      "properties": {
        "description": {
          "type": "string",
          "example": "Made in CB-TB"
        },
        "installMonAgent": {
          "type": "string",
          "description": "InstallMonAgent Option for CB-Dragonfly agent installation ([yes/no] default:no)",
          "default": "no",
          "enum": [
            "yes",
            "no"
          ],
          "example": "no"
        },
        "label": {
          "type": "object",
          "description": "Label is for describing the object by keywords"
        },
        "name": {
          "type": "string",
          "example": "mci01"
        },
        "systemLabel": {
          "type": "string",
          "description": "SystemLabel is for describing the mci in a keyword (any string can be used) for special System purpose",
          "example": ""
        },
        "vm": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "commonImage": {
                "type": "string",
                "description": "CommonImage is field for id of a image in common namespace",
                "example": "ubuntu18.04"
              },
              "commonSpec": {
                "type": "string",
                "description": "CommonSpec is field for id of a spec in common namespace",
                "example": "aws+ap-northeast-2+t2.small"
              },
              "connectionName": {
                "type": "string",
                "description": "if ConnectionName is given, the VM tries to use associtated credential.\nif not, it will use predefined ConnectionName in Spec objects"
              },
              "description": {
                "type": "string",
                "example": "Description"
              },
              "label": {
                "type": "object",
                "description": "Label is for describing the object by keywords"
              },
              "name": {
                "type": "string",
                "description": "VM name or subGroup name if is (not empty) && (> 0). If it is a group, actual VM name will be generated with -N postfix.",
                "example": "g1-1"
              },
              "rootDiskSize": {
                "type": "string",
                "description": "\"default\", Integer (GB): [\"50\", ..., \"1000\"]",
                "default": "default",
                "example": "default, 30, 42, ..."
              },
              "rootDiskType": {
                "type": "string",
                "description": "\"\", \"default\", \"TYPE1\", AWS: [\"standard\", \"gp2\", \"gp3\"], Azure: [\"PremiumSSD\", \"StandardSSD\", \"StandardHDD\"], GCP: [\"pd-standard\", \"pd-balanced\", \"pd-ssd\", \"pd-extreme\"], ALIBABA: [\"cloud_efficiency\", \"cloud\", \"cloud_essd\"], TENCENT: [\"CLOUD_PREMIUM\", \"CLOUD_SSD\"]",
                "default": "default",
                "example": "default, TYPE1, ..."
              },
              "subGroupSize": {
                "type": "string",
                "description": "if subGroupSize is (not empty) && (> 0), subGroup will be generated. VMs will be created accordingly.",
                "default": "1",
                "example": "3"
              },
              "vmUserPassword": {
                "type": "string"
              }
            }
          }
        }
      }
    },
    "path_params": {
      "required": [
        "nsId"
      ],
      "properties": {
        "nsId": {
          "type": "string",
          "description": "Namespace ID",
          "default": "default"
        }
      }
    },
    "query_params": {
      "properties": {
        "option": {
          "type": "string",
          "description": "Option for MCI creation",
          "enum": [
            "hold"
          ]
        }
      }
    }
  },
  "created_at": "2024-11-01T14:16:21.541043563+09:00",
  "updated_at": "2024-11-01T17:03:37.326385637+09:00",
  "is_example": true
}
```

</details>

## SMTP 

### 1. Add SMTP info
file path : /_airflow/docker-compose.yml 

modify docker-compose.yml file and enter your smtp info.

gmail example : https://support.google.com/a/answer/176600?hl=en


```
...
    airflow-server:
        environment:
            AIRFLOW__SMTP__SMTP_HOST: 'smtp.gmail.com'
            AIRFLOW__SMTP__SMTP_USER: 'yourEmail@gmail.com'
            AIRFLOW__SMTP__SMTP_PASSWORD: 'wtknvaprkkwyaurd'
            AIRFLOW__SMTP__SMTP_PORT: 587
            AIRFLOW__SMTP__SMTP_MAIL_FROM: 'yourEmail@gmail.com'
...
```
### 2. Modify mail.py 
file path : /_airflow/airflow-home/dags/mail.py

Modify the recipient's email address in the email_task.

```
...
    email_task = EmailOperator(
        task_id='send_email',
        to='Your Email@example.com',
        subject='DAG 상태 보고서',
        ...
    )
...
```

### 3. Add taskComponent 
Add trigger_email task component at the bottom of the workflow to receive email alarms.

```
...
         {
           "name": "trigger_email",
           "task_component": "trigger_email",
           "request_body": "",
           "path_params": {},
           "dependencies": [
             "{$Pre_taskName}"
           ]
         }

...
```

## GET task log 
### 1. GET workflow RunId
[GET] /workflow/{wfId}/runs
![image](https://github.com/user-attachments/assets/27fbaf5f-c52d-4d04-b599-ef2eac9e76de)

### 2. GET taskId and task_Try_Num
[GET] /workflow/{wfId}/workflowRun/{wfRunId}/taskInstances
![image](https://github.com/user-attachments/assets/d893cc1a-2cbd-417c-a19d-a650aaca7f6e)

### 3. GET execution task log
[GET] /workflow/{wfId}/workflowRun/{wfRunId}/task/{taskId}/taskTryNum/{taskTyNum}/logs
![image](https://github.com/user-attachments/assets/347babf5-df32-4fe0-82e0-f0e111c333d1)


## Health-check

Check if CM-Cicada is running

```bash
curl http://127.0.0.1:8083/cicada/readyz

# Output if it's running successfully
# {"message":"CM-Cicada API server is ready"}
```

## Check out all APIs
* [Cicada APIs (Swagger Document)](https://cloud-barista.github.io/cb-tumblebug-api-web/?url=https://raw.githubusercontent.com/cloud-barista/cm-cicada/main/pkg/api/rest/docs/swagger.yaml)
