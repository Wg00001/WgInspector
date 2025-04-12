docker build -t wginspector:latest .

docker run -d -p 9999:9999 --name wginspector wginspector:latest

docker run wginspector
