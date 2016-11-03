## Build Instructions



### Windows

1. Visit [golang.org/dl](https://golang.org/doc/install?download=go1.7.3.windows-amd64.msi) to download a copy of the Go installer for Windows.
2. Run the MSI installer file downloaded in the previous step.
3. Close and reopen any command prompts for changes to environmental variables to take effect.
4. Change to the workspace directory in which you'd like to clone Verbatim.
5. Set your `GOPATH` environmental variable to your workspace directory.

	```{batchfile}
	set GOPATH=C:\workspace_dir
	```
6. Open Git Bash and clone Verbatim. If you don't have Git installed, you can download Verbatim's [source as a zip](https://github.com/0x7fffffff/verbatim/archive/master.zip) file instead.

	```{shell}
	git clone https://github.com/0x7fffffff/verbatim
	```

7. Change into the verbatim directory.
8. Run the following command to fetch all of Verbatim's dependencies.

	```{shell}
	go get
	```

9. Build Verbatim in production mode.

	```{shell}
	go build -tags prod
	```
10. Run the newly created `Verbatim.exe` executable.


### macOS

Visit [golang.org/dl](https://golang.org/dl/)

### Linux

Visit [golang.org/dl]()

### Other Platforms

If you're running on a platform not listed above, there's still a chance that you can build Verbatim. Visit the [installing Go from source](https://golang.org/doc/install/source?download=go1.7.3.src.tar.gz) page on [golang.org](https://golang.org) for specific instructions on how to build Go on your platform.

