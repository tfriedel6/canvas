As of this writing, gomobile does not support go modules. In this case this project can only be compiled while it is in the GOPATH/src directory.

The go bindings are generated with the ```gomobile bind -target android``` command, which results in a .aar and a .jar file. These should be placed in the CanvasAndroidExample/app/libs directory, and then the project should compile.
