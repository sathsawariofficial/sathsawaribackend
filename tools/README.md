# start log cleanup
0 0 * * * /path/to/clean_logs.sh >> /var/log/cleanup.log 2>&1
chmod +x /path/to/clean_logs.sh

crontab -e // create
crontab -l // verify
