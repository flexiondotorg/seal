# Seal

seal is rtmp server written by go language, main refer to rtmp server open source https://github.com/ossrs/srs

## Usage
* build

  download https://github.com/calabashdad/seal to ```go path```, run ```go build```

  you can also use cross platform build, like build a linux version if you are on mac, run ```cross_platform_linux```

* run console mode

  ```./seal -c seal.yaml```
* run daemon mode

  ```nohup ./seal -c seal.yaml &```
* mock stream publish

  ```ffmpeg -f video4linux2 -s 1280x720 -i /dev/video4 -f flv -b:v 1024k -metadata streamName="Webcam" rtmp://127.0.0.1/live/test;```

* use vlc play

  rtmp ```rtmp://127.0.0.1/live/test```

  hls  ```http://127.0.0.1:35418/live/test.m3u8```
  
  http-flv ```http://127.0.0.1:35418/live/test.flv```

## platform
  go is cross platform 
* linux
* mac
* windows

## support
* rtmp protocol (h264 aac)
* hls (include http server)
* http-flv (include http server)

## plan to support
* h265
* transcode(audio to aac)
* http stats query
* video on demand
* video encry
* auth token dynamicly
* mini rtmp server in embed device
