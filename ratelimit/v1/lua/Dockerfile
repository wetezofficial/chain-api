FROM nickblah/lua:5.1-luarocks-bionic

RUN apt-get update && \
    apt-get install -y build-essential redis-tools vim && \
    luarocks install luacov && \
    luarocks install busted && \
    luarocks install redis-lua

CMD [ "tail", "-f", "/dev/null" ]
