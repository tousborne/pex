Pex challenge.

### Ideas
The idea behind the program design is to limit the amount of memory the program can consume via a
buffer of images to be processed and then running a goroutine per image to process the top three
colors and allowing the go runtime to hammer the available CPUs.  Anecdotally, using the provided
limits (512M and 1CPU) in a cgroup yielded performance results that indicated the CPU was a severe
bottleneck.  For a much more scalable solution, having finer grain control such as seperate
downloading and processing routines whose parameters could be independently tuned would be more
ideal.  However, the CPU is such a bottleneck that having a single tunable parameter of the number
of images to concurrently process did not result in any noticable performance changes as the
network and memory usage were insignificant compared to CPU usage.  If 1-2 orders of magnitude more
CPUs were thrown at the problem relative to the available memory and network resources, then the
above optimization would begin to make sense.

Possible optimizations include:
* Parsing an evenly divided subset of every image (such as every 10th pixel), then running a
	statistical analysis of the processed pixels to determine if further parsing rounds are needed
	of a larger subset of the image.
* Breaking up the downloading and processing into seperate goroutines.
  * Anecdotaly this did not make a difference and so was abanoned, but I do believe that it would
    be useful in a highly scaled environment.


### Usage
```
go build
./pex --in <input file> --out <output file> --size <int>
```

The input and output files must be specified (`--in` and `--out`).  The size (`--size`) is optional,
but is recommended as it sets the number of images to process concurrently and the default is 1.


### Notes
The application only supports JPEG and PNG image formats and only NRGBA and YCrCb color models as
this seemed to be the only image types in the sample set.


### Profiling
#### CPU usage
As can be seen in the two images below, the majority of the CPU time is spent parsing the individual
pixels in each image and decoding the images.  I believe that implementing an image format library
is beyond the scope of this challenge and I couldn't come up with a useful image analysis that would
allow parsing fewer then every pixel in every image.

![small file cpu usage slice](https://raw.githubusercontent.com/tousborne/pex/master/profiling/cpu_slice_small_images.gif)
![normal file cpu usage slice](https://raw.githubusercontent.com/tousborne/pex/master/profiling/cpu_slice_normal_images.gif)


#### Memory usage
As can be seen in the image below the vast majority of the memory usage was in the libraries
responsible for decoding the images and I believe optimizing an image format library is beyond the
scope of this challenge.

![normal file memory usage slice](https://raw.githubusercontent.com/tousborne/pex/master/profiling/memory_usage_slice_normal_images.gif)
