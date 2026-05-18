docker build \
-f /home/raotalha/Code/PersonalCode/sathsawaribackend/Dockerfile \
-t rideshare .

docker save rideshare | gzip > rideshare.tar.gz

docker run \
--network host \
-v $(pwd)/docker/configuration.json:/configuration.json \
-v $(pwd)/docs:/docs \
-v $(pwd)/docker/logs:/logs \
-p 5000:5000 rideshare
