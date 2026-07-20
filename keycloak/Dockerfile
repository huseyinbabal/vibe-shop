# Realm import baked into the image so the compose file needs no bind mounts
# (Dokploy clones sources into its own container; bind paths would not resolve
# on the host daemon).
FROM quay.io/keycloak/keycloak:26.3
COPY vibe-shop-realm.json /opt/keycloak/data/import/vibe-shop-realm.json
