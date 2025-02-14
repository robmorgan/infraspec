FROM node:20-slim

WORKDIR /app

COPY package.json package.json
COPY index.js index.js

RUN npm install

EXPOSE 3000

CMD ["node", "index.js"]
