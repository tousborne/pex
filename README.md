Pex challenge.

### Ideas
My assumption is that memory is the limiting factor in this situation and therefore I approached the situation with the goal of maximizing a certian amount of memory usage.  To acheive that, I specify a buffer size that limits the number of images that are held in memory and being processed.  It would have been better to implement memory and goroutine pools, but I think that's a little outside the scope of this challenge, so instead I played with channels.

Possible optimizations include:
* Break the processing of individual images up into a divide and conquer approach
* Breaking up the downloading and processing into seperate goroutines
* Maintaining the in-use amount of memory and optimizing the number of downloading and processing goroutines to always have a full "pipeline"
* Need to run an profiler to see if there are any obvious things I'm missing

### Usage
```
go build
./pex --in <input file> --out <output file> --size <int>
```

The input and output files must be specified (`--in` and `--out`).  The size (`--size`) is optional, but is recommended as it sets the number of images to process concurrently and the default is 1.


### Notes
There are probably some less than ideal assumptions made in order to keep the submission simple.  Those are hopefully noted in comments wherever they exist.

The application only supports JPEG and PNG image formats and only NRGBA and YCrCb color models.
