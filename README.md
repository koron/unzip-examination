# unzip examination

## Preparation

Copy golang windows binary distribution (zip) as test.zip

## How to run

```console
$ go build main4.go
$ ./main4 -f -t -p 8

# for clean up
$ rm -rf outdir
```

## Result

* `-f -t -p 8`
    * 12.442168s
    * 11.6765334s
* `-f -p 8`
    * 11.6692459s
    * 12.3265582s
* `-f -t -p 1` : 52.1837447s
* `-f -t -p 0` : 11.8378314s
* `-f -t -p 4` : 14.9077009s
* `7za x` : 18.996s
* `unzip -q` : 1m25.019s
