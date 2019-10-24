FROM ubuntu:18.04

RUN apt-get update && apt-get -y upgrade
RUN apt-get install -y git build-essential autoconf bison re2c libxml2-dev zlib1g-dev nginx

RUN git clone https://github.com/php/php-src

# checkout the fix
RUN git -C php-src checkout ab061f95ca966731b1c84cf5b7b20155c0a1c06a

# checkout the commit previous to the fix
RUN git -C php-src checkout HEAD~1

# build php-fpm
RUN cd php-src && ./buildconf --force && ./configure --enable-fpm --without-pear && make -j4 && make install

COPY php-fpm.conf /usr/local/etc/
COPY nginx.server.conf /etc/nginx/sites-enabled/default
COPY script.php /var/www/html/script.php
COPY entrypoint /

CMD /entrypoint