Pex challenge.

### Ideas
The idea behind the program is to be able to tune the number of downloading and executing
goroutines in order to maximize the throughput on a given set of hardware.  The
downloading threads push their downloaded image data to a buffer that the parsing threads
can pull from.  The downloading threads can also form an ad-hoc buffer as they will hold
the image data in memory if the buffer pipeline is full. Anecdotally, using the provided
limits of 512M and 1 CPU in a CGroup yielded performance results that indicated the CPU
was a severe bottleneck, but I have a pretty good internet connection and my test data may
not have been representative of the real world.

Possible optimizations include:
* Parsing an evenly divided subset of every image (such as every 10th pixel), then running
  a statistical analysis of the processed pixels to determine if further parsing rounds
  are needed of a larger subset of the image.


### Usage
```
go build
./pex --in <input file> --out <output file> --down <int> --exec <int>
```

The input and output files must be specified (`--in` and `--out`).  The "down" and "exec"
(`--down` and `--exec`) flags are optional, but it is recommended to set them as the tune
the number of downloading and image parsing goroutines and therefore control memory and
cpu usage.  Be careful though, as high numbers can easily eat through the system
resources.

Note: I tended to find that using around double the number of downloading threads to
execution threads maximized throughput, with 300 downloading goroutines and 150 executing
goroutines being the performance sweet spot on my modern laptop.


### Notes
The application only supports JPEG and PNG image formats and only NRGBA and YCrCb color
models as this seemed to be the only image types in the sample set.


### Profiling
#### CPU usage
As can be seen in the two images below, the majority of the CPU time is spent parsing the
individual pixels in each image and decoding the images.  I believe that implementing an
image format library is beyond the scope of this challenge and I couldn't come up with a
useful image analysis that would allow parsing fewer then every pixel in every image.

![small file cpu usage slice](https://raw.githubusercontent.com/tousborne/pex/master/profiling/cpu_slice_small_images.gif)
![normal file cpu usage slice](https://raw.githubusercontent.com/tousborne/pex/master/profiling/cpu_slice_normal_images.gif)


#### Memory usage
As can be seen in the image below the vast majority of the memory usage was in the
libraries responsible for decoding the images and I believe optimizing an image format
library is beyond the scope of this challenge.

![normal file memory usage slice](https://raw.githubusercontent.com/tousborne/pex/master/profiling/memory_usage_slice_normal_images.gif)
