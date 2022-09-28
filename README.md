## start
```sh
docker container run -d --name buildkitd --privileged moby/buildkit:latest
export BUILDKIT_HOST=docker-container://buildkitd
```

## run
```
go run main.go | buildctl build --local context=. --output type=image,name=test:llb
go run main.go | buildctl build --local context=. --output type=local,dest=output

```