FROM bitnami/kubectl:latest

USER root

# Add any additional packages or configurations here
RUN apt update && apt -y install python3 wget
RUN wget https://raw.githubusercontent.com/rabbitmq/rabbitmq-server/v3.12.x/deps/rabbitmq_management/bin/rabbitmqadmin

RUN chmod +x rabbitmqadmin
RUN mv ./rabbitmqadmin /usr/bin/

USER 1001

CMD ["/bin/bash"]