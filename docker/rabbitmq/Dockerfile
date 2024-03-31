FROM rabbitmq:3.12.12-alpine

#RUN rabbitmq-plugins enable --offline rabbitmq_management
#RUN rabbitmq-plugins enable --offline rabbitmq_management rabbitmq_shovel rabbitmq_shovel_management

RUN mkdir -p /etc/rabbitmq/conf.d
RUN chown -R rabbitmq:rabbitmq /etc/rabbitmq/conf.d
COPY 20-logging.conf /etc/rabbitmq/conf.d

EXPOSE 5672