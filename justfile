
build :
    go build ./cmd/spigot

[working-directory: 'examples/gotext']
run : build
    rm -f /var/tmp/spigot_hotness*.log
    ../../spigot -c ./gotext.yml

[working-directory: 'cmd/spigot']
original : build
    rm -f /var/tmp/spigot_*.log
    ../../spigot -c ./spigot.yml
