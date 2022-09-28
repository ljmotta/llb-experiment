package main

import (
	"context"
	"fmt"
	"os"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/imagemetaresolver"
)

func main() {
	dt, err := createLLBOutputFiles(createLLBBaseImage()).Marshal(context.TODO(), llb.LinuxAmd64)

	if err != nil {
		panic(err)
	}
	llb.WriteTo(dt, os.Stdout)
}

func createLLBBaseImage() llb.State {
	var extensions = false
	var extensionList = "quarkus-jsonb"
	var workflowName = "test"
	var containerRegistry = "quay.io"
	var containerGroup = "lmotta"
	var containerName = "my-proj"
	var containerTag = "llb"

	opts := []llb.ImageOption{llb.LinuxAmd64}
	opts = append(opts, imagemetaresolver.WithDefault)

	base := llb.Image("quay.io/lmotta/kn-workflow:2.10.0.Final", opts...)
	base = base.Dir("/tmp/kn-plugin-workflow")

	if extensions {
		base = base.
			Run(
				llb.Shlex(fmt.Sprintf("./mvnw quarkus:add-extension -Dextensions=%s", extensionList)),
			).
			Root()
	} else {
		base = base.
			Run(
				llb.Shlex("echo \"WITHOUT ADDITIONAL EXTENSIONS\""),
			).
			Root()
	}

	base = base.File(llb.Copy(llb.Local("context"), "./workflow.sw.json", "./src/main/resources/workflow.sw.json"))
	// TODO: add check
	// base.File(llb.Copy(llb.Local("context"), "application.properties", "./src/main/resources/application.properties"))

	base = base.
		Run(
			llb.Shlex(
				"./mvnw package" +
					" -Dquarkus.kubernetes.deployment-target=knative" +
					fmt.Sprintf(" -Dquarkus.knative.name=%s", workflowName) +
					fmt.Sprintf(" -Dquarkus.container-image.registry=%s", containerRegistry) +
					fmt.Sprintf(" -Dquarkus.container-image.group=%s", containerGroup) +
					fmt.Sprintf(" -Dquarkus.container-image.name=%s", containerName) +
					fmt.Sprintf(" -Dquarkus.container-image.tag=%s", containerTag),
			)).
		Root()

	return base
}

func createLLBOutputFiles(base llb.State) llb.State {
	outputFiles := llb.Scratch()
	var CopyOptions = &llb.CopyInfo{
		FollowSymlinks:      true,
		CopyDirContentsOnly: true,
		AttemptUnpack:       false,
		CreateDestPath:      true,
		AllowWildcard:       true,
		AllowEmptyWildcard:  true,
	}
	outputFiles = outputFiles.File(
		llb.Copy(base, "/tmp/kn-plugin-workflow/", ".", CopyOptions),
	)
	return outputFiles
}

func createLLBRunnerImage(base llb.State) llb.State {
	runner := llb.Image("opendjdk:11")
	var CopyOptions = &llb.CopyInfo{
		FollowSymlinks:      true,
		CopyDirContentsOnly: true,
		AttemptUnpack:       false,
		CreateDestPath:      true,
		AllowWildcard:       true,
		AllowEmptyWildcard:  true,
	}
	runner = runner.File(
		llb.
			Copy(base, "/tmp/kn-plugin-workflow/target/quarkus-app/lib/", "/runner/lib/", CopyOptions).
			Copy(base, "/tmp/kn-plugin-workflow/target/quarkus-app/*.jar", "/runner/", CopyOptions).
			Copy(base, "/tmp/kn-plugin-workflow/target/quarkus-app/app/", "/runner/app/", CopyOptions).
			Copy(base, "/tmp/kn-plugin-workflow/target/quarkus-app/quarkus/", "/runner/quarkus/", CopyOptions),
	)
	return runner
}

func createLLBDevImage(base llb.State) llb.State {
	dev := llb.Image("opendjdk:11")
	var CopyOptions = &llb.CopyInfo{
		FollowSymlinks:      true,
		CopyDirContentsOnly: true,
		AttemptUnpack:       false,
		CreateDestPath:      true,
		AllowWildcard:       true,
		AllowEmptyWildcard:  true,
	}
	dev = dev.File(
		llb.
			Copy(base, "/root/.m2/", "/root/.m2/", CopyOptions).
			Copy(base, "/tmp/", "/tmp/", CopyOptions),
	)

	dev.Dir("/tmp/kn-plugin-workflow/")
	return dev
}
