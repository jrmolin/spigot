
[working-directory: 'cmd/spigot']
build :
    go build .

[working-directory: 'examples/gotext']
run : build
    rm -f /var/tmp/spigot_hotness*.log
    ../../cmd/spigot/spigot -c ./gotext.yml

[working-directory: 'cmd/spigot']
original : build
    rm -f /var/tmp/spigot_*.log
    ../../cmd/spigot/spigot -c ./spigot.yml
