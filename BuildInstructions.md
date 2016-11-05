## Build Instructions


Listed below are build instructions for [Windows](#windows), [macOS](#macos) & [Linux](#linux). Verbatim is capable of running any platform/architecture that both [Go](https://golang.org) and [GCC](https://gcc.gnu.org) are available on. If your platform isn't listed below, please visit [golang.org/doc/install/source](https://golang.org/doc/install/source) and [gcc.gnu.org/install](https://gcc.gnu.org/install/) for detailed install instructions on building Go and GCC from source for your platform.

If you're using one of the listed platforms below, there may be a precompiled Verbatim binary available on [Verbatim's release page](https://github.com/0x7fffffff/verbatim/releases). If downloading a binary, be sure to download the latest stable version.

**Note:** When Go is referenced in this document, it is assumed to mean Go version 1.7.3 or higher.

### Windows

1. Visit [golang.org/dl](https://golang.org/dl) to download a copy of the Go installer for Windows.
2. Run the MSI installer file downloaded in the previous step.
3. Close and reopen any command prompts for changes to environmental variables to take effect.
4. Change to the (workspace) directory in which you'd like to clone Verbatim.
5. Set your `GOPATH` environmental variable to your workspace directory.

	```{batchfile}
	set GOPATH=C:\workspace_dir
	```
6. Open Git Bash and clone Verbatim. If you don't have Git installed, you can either [install it](https://git-scm.com/download/), or download Verbatim's [source as a zip](https://github.com/0x7fffffff/verbatim/archive/master.zip) file instead.

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

1. Visit [golang.org/dl](https://golang.org/dl/) to download a copy of the Go installer for macOS.
2. Run the pkg installer file downloaded in the previous step.
3. Close and reopen any terminal windows for changes to environmental variables to take effect.
4. Change to the (workspace) directory in which you'd like to clone Verbatim.
5. Set your `GOPATH` environmental variable to your workspace directory.
	
	```{shell}
	export GOPATH=workspace_dir
	```

6. Clone Verbatim. If you don't have Git installed, you can either [install it](https://git-scm.com/download/), or download Verbatim's [source as a zip](https://github.com/0x7fffffff/verbatim/archive/master.zip) file instead.

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

10. Run the newly created `verbatim` executable with `./verbatim`.
11. If step 10 fails, you may need to make the Verbatim binary executable. This can be done with the following command. Then try step 10 again.

	```{shell}
	chmod +x verbatim
	```

### Linux

Visit [golang.org/dl]()

### Other Platforms


