## Structure

The structure is based on ADP Reference Application.
[More details about this pattern](https://github.com/golang-standards/project-layout).


### Directories

#### - /charts

This directory contains Helm charts.

#### - /ci

This directory contains CI configuration files.

#### - /cmd

###### Overview:
Main functions for this project.
The directory name for each application should match the name of the executable you want to have (e.g., /cmd/myapp).
Don't put a lot of code in the application directory. If you think the code can be imported and used in other projects,
then it should not live in this directory. If the code is not reusable or if you don't want others to reuse it,
put that code in the /internal directory.

It's common to have a small main function that imports and invokes the code from the /internal directory and nothing else.


###### Implementation details:
The main logic of the microservice is implemented in the file cmd/go-template-service/server.go.
At the end of the file, the main() function is located.

#### - /internal

###### Overview:
Private application and library code. This is the code you don't want others importing in their applications or
libraries. Note that this layout pattern is enforced by the Go compiler itself.
See the Go 1.4 release notes for more details. Note that you are not limited to the top level internal directory.
internal/ is a special directory name recognised by the go tool which will prevent one package from being imported by
another unless both share a common ancestor.
You can have more than one internal directory at any level of your project tree.

###### Implementation details:
github.com/fsnotify/fsnotify library is used

In config.go, the structure and the values are specified for the configuration holding object
(in type Config struct ... and func getConfig() *Config ..., respectively),
aided by helper functions getOsEnvInt() and getOsEnvString() that take the specific
values from the environment - along with default values if they are not present.
The configuration object is ready to be used.

#### - /test

###### Overview:
Additional external test apps and test data. Feel free to structure the /test directory anyway you want.

### Directories You Shouldn't Have

#### - /src

Some Go projects do have a src folder, but it usually happens when the devs came from the Java world where it's a common pattern.