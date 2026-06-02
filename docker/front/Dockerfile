FROM php:8.2-fpm-alpine

WORKDIR /var/www/front

# Passa variáveis de ambiente do docker-compose para os workers do PHP-FPM
COPY env.conf /usr/local/etc/php-fpm.d/env.conf

CMD ["php-fpm"]
