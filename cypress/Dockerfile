FROM cypress/included

RUN mkdir /app
WORKDIR /app
COPY . /app
RUN npm install

ENTRYPOINT ["tail", "-f", "/dev/null"]
