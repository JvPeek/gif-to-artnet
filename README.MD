# gif-to-artnet

gets images (gif, jpeg, png) via http, resizes them to your LED matrix and displays them via artnet for a certain amount of time. Can do animated gifs. 

Useful for injecting short animations onto a WLED matrix that will continue it's normal effects after a few seconds.

## ToDo:
- Read more than just .gif images
- Send them via ArtNet and remove them from the queue
- Create some kind of config file to set things like destination IP, matrix resolution, display duration, possible "active now"-webhooks
