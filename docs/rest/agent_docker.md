# Run the agent as docker container

## Build the Agent
Build the docker image for `agent-rest` by running following make target from project root directory. 

`make agent-rest-docker`

## Run the Agent
Above target will build docker image `docker.pkg.github.com/trustbloc/agent-sdk/agent-sdk-rest` which can be used to start agent by running command as simple as 

```
 docker run docker.pkg.github.com/trustbloc/agent-sdk/agent-sdk-rest start [flags] 
```

Details about flags can be found [here](agent_cli.md#Agent-Parameters)