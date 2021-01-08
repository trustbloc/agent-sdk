# Run the agent as docker container

## Build the Agent
Build the docker image for `agent-server` by running following make target from project root directory. 

`make agent-server-docker`

## Run the Agent
Above target will build docker image `ghcr.io/trustbloc/agent-sdk-server` which can be used to start agent by running command as simple as 

```
 docker run ghcr.io/trustbloc/agent-sdk-server start [flags] 
```

Details about flags can be found [here](agent_cli.md#Agent-Parameters)