## Image merger and resizer

This repository contains a Go application that can merge and resize images on the commandline.
Currently it supports the following image formats:

* BMP
* JPEG
* GIF
* PNG

Currently, this application has not been fully tested, as it was written in an afternoon.


### Compiling and running

Compile the application by running

```bash
go build image_merger_and_resizer.go
```

Now you can scale or merge images.

Scale an image by calling

```bash
./image_merger_and_resizer <input_file> scale <scale_factor> [output_file]
```

Merge two images by calling

```bash
./image_merger_and_resizer <input_file> merge[:offsetX][:offsetY] <merge_file> [output_file]
```

Note that if you specify the `.bmp`, `.jp(e)g`, `.gif` or `.png` file extensions, this application will encode
the output in the corresponding encoding.