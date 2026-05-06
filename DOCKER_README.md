# Docker build & Back4App Container notes

Quick steps to build and run locally:

```bash
# Build image
docker build -t backend-skripsi:latest .

# Run with PORT env (app prefers `PORT`, then `APP_PORT`)
docker run -e PORT=8080 -p 8080:8080 backend-skripsi:latest
```

Back4App Container (or other container hosts):

- Connect your repository or push an image to a registry. Back4App can build from your repo using this `Dockerfile`.
- Configure required environment variables in the Back4App dashboard (do NOT bake secrets into the image):
  - `JWT_SECRET`, DB connection settings, Cloudinary credentials, etc.
- The app reads `PORT` then `APP_PORT` then defaults to `8080`.

Notes:

- Do not commit `.env` into the repo; use platform environment variables.
- If you need a smaller image, we can switch to a distroless final stage.
