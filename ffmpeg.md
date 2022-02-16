youtube-dl -f 137+140 "https://www.youtube.com/watch?v=xxxx"
ffmpeg -y -i xxxx.mp4 -ss 4:05 -t 1:27 -c copy 1.mp4
ffmpeg -y -i xxxx.mp4 -ss 1:00:50 -t 2:36 -c copy 2.mp4
ffmpeg -y -f concat -safe 0 -i <(cat << EOF
file '$(pwd)/1.mp4'
file '$(pwd)/2.mp4'
EOF
) -c copy output.mp4
