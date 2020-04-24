# go-mesos-executor
executor for containerd

containerd version should be 1.3.x

an example json:
{
  "id": "test1",
  "cpus": 0.2,
  "mem": 256,
  "cmd": "--image docker.io/library/redis:alpine --namespace test --command \"sleep 100\"",
  "instances": 1,
  "executor": "./go-mesos-executor",
   "fetch": [
    {
      "uri": "http://172.31.35.88:9900/go-mesos-executor",
      "extract": true,
      "executable": false,
      "cache": false
    },
    {
      "uri": "http://172.31.35.88:9900/config.yaml",
      "extract": true,
      "executable": false,
      "cache": false
    }
  ]
}

optional parameter: namespace and command
required parmeter: image
