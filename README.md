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

