## Image merger and resizer

This repository contains a Go application that can merge and resize images on the commandline.
Currently it supports the following image formats:

* BMP
* JPEG
* GIF
* PNG

There are two example run configurations in the `.run-local` directory.


### Current state

The application has been tested to successfully work with PNG input files. The other formats listed are supported,
but have not been tested at the time of writing.
File formats that will be added in the near future:

* TIFF
* WebP

Please file an issue or email me at [ahcwolvers@gmail.com](mailto:ahcwolvers@gmail.com) if there are other file formats
you would like to see supported, or have any questions or suggestions about this application.

&nbsp;

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

Also note that if you do not specify an output file, this application
will print directly to stdout.