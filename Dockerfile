FROM vault:1.5.4

RUN mkdir /tmp/vault-plugins
COPY dist/vault-launchdarkly-secrets-engine_linux_amd64/vault-launchdarkly-secrets-engine /tmp/vault-plugins
RUN echo 'plugin_directory = "/tmp/vault-plugins"' >> /vault/config/plugin.hcl
#RUN echo 'default_lease_ttl="60s"' >> /vault/config/plugin.hcl
#RUN echo 'max_lease_ttl="1200s"' >> /vault/config/plugin.hcl
