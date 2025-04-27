package react

import (
	"fmt"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(facts core.Facts) (string, error) {
	dockerfile := fmt.Sprintf(`FROM node:14
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE %d
CMD ["npm", "start"]`, facts.Port)

	return dockerfile, nil
}
