# HTML Analyzer Readme

## Description

This project implements a web application written in Go that analyzes a provided webpage URL. The application fetches the webpage content and processes it to extract various details. The analysis results are then displayed to the user, including:

* **HTML Version:** The version of HTML used by the webpage.
* **Page Title:** The title of the webpage as specified in the `<title>` tag.
* **Headings:** The number of headings for each level (e.g., H1, H2, H3) present in the document.
* **Links:** 
    * The total number of internal links (pointing to resources within the same domain).
    * The total number of external links (pointing to resources outside the domain).
    * The number of inaccessible links (broken links).
* **Login Form:** Whether the webpage contains a login form element.

## Note
This repository contains two binaries: an API and a website.

## Dependencies

To build and run the applications, ensure you have the following dependencies installed:

- [Go](https://golang.org/doc/install)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Building the Applications

Use these commands to build the applications:

### API
```bash
make build-api
```

### Website
```bash
make build-website
```

### Docker
```sh
make docker-build
```

## Running the Applications

Run the applications using these commands:

### Docker Compose

This will start both applications: the API on port 8080 and the website on port 3000.
```sh
make docker-compose-up
```

### API
```sh
make run-api
```

### Website
```sh
make run-website
```

## Development

### Website

The website uses the libraries [templ](https://github.com/a-h/templ/tree/main) and [HTMX](https://htmx.org/) to generate the final HTML files.

In case you need to regenerate any template, you can run the command
```sh
make go-generate
```

## Testing

Besides the unit testing, the project includes user acceptance testing located in the folder `/cmd/acceptance_test/`.

The user acceptance testing uses the libraries [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/)

If you want to execute only the acceptance tests, run the command
```sh
make test-ginkgo
```


## Deployment

The applications can be deployed using the generated Docker images.

### API

The API container has an optional environment variable named `PORT`, which defines the port on which it will listen for requests.

**Example:**

```bash
docker run -p 8080:8080 -e PORT=8080 <docker image>
```

### Website

The website container requires an environment variable named `SERVER_HOST` to be set, which determines the hostname used by the api server.

**Example:**

```bash
docker run -p 3000:3000 SERVER_HOST=api-host.com --entrypoint "/website" <docker image>
```

## Possible Improvements

- Currently, only HTML version 5 and 4 are supported. This could be expanded to older versions and distinct between loose and strict HTML files.

- To verify if the html file has a login form, only it's checked if the page contains a form with an input field with type password inside. That means that register forms could also be misinterpreted as Login forms.

- Add metrics

- Add testing to website project.

- Honestly, it's my first time working with HTMX and Go website interfaces. Definitely, the whole website project can be improved.