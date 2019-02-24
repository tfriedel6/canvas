As of this writing, gomobile does not support go modules. In this case this project can only be compiled while it is in the GOPATH/src directory.

Run this command:

gomobile bind -target ios -tags ios

Then add the resulting Example.framework into the Xcode project, and it should compile and run from there
