#!/bin/sh
set -e

cd /var/www/back

composer install --no-interaction --prefer-dist --optimize-autoloader

exec php-fpm
