# Pix

Turn photos into abstract art.

![Road in the Winter Forest by Olga Malamud Pavlovich](img/winter.png)

Install the command-line tool with `go get`:

```
go get -u github.com/yurivish/pix/cmd/pix
```

Run it like so:

```
pix -in picture.jpg
```

Generate multiple outputs by sweeping the parameter space:

```
pix -in picture.jpg -sweep
```

Pix is capable of generating 8,000×8,000 outputs in around a minute. 

Since the placement process involves _n_ nearest-neighbor searches, where _n_ is the number of pixels in the output image, the time taken depends significantly on the color distribution and placement order. These affect the shape of the frontier and the size of the search tree.