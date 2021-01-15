{
  config:: {
    hostname: null,
    docker_tag: 'latest',
    dev: '',
    commit_sha: '',

    // derived
    nameSuffix: if self.dev != '' then '-' + self.dev else '',
    hostPrefix: if self.dev != '' then self.dev + '.' else '',
  },
  configure(overlay={}, hostname=null, docker_tag=null, dev=null, commit_sha=null)::
    self + overlay + {
      config+: std.prune({
        hostname: hostname,
        docker_tag: docker_tag,
        dev: dev,
        commit_sha: commit_sha,
      }),
    },

  manifest: [
    $.namespace,
    $.service,
    $.deployment,
    if $.config.hostname != null then $.ingress,
  ],
  namespace:: {
    apiVersion: 'v1',
    kind: 'Namespace',
    metadata: {
      name: 'foxtrot',
    },
  },
  service:: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      namespace: 'foxtrot',
      name: 'foxtrot' + $.config.nameSuffix,
      labels: {
        app: 'foxtrot',
        dev: $.config.dev,
      },
    },
    spec: {
      ports: [{ name: 'http', port: 8080 }],
      selector: {
        app: 'foxtrot',
        dev: $.config.dev,
      },
    },
  },
  deployment:: {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: {
      namespace: 'foxtrot',
      name: 'foxtrot' + $.config.nameSuffix,
      labels: {
        app: 'foxtrot',
        dev: $.config.dev,
      },
    },
    spec: {
      selector: {
        matchLabels: {
          app: 'foxtrot',
          dev: $.config.dev,
        },
      },
      template: {
        metadata: {
          labels: {
            app: 'foxtrot',
            dev: $.config.dev,
          },
          annotations: {
            commit_sha: $.config.commit_sha,
          },
        },
        spec: {
          containers: [
            {
              local policy(tag) = if tag == 'latest' || std.startsWith(tag, 'pr') then 'Always' else 'IfNotPresent',
              image: 'foxygoat/foxtrot:%s' % $.config.docker_tag,
              imagePullPolicy: policy($.config.docker_tag),
              name: 'foxtrot',
              ports: [{ containerPort: 8080, name: 'http', protocol: 'TCP' }],
              env: [{
                name: 'FT_AUTH_SECRET',
                valueFrom: { secretKeyRef: { key: 'authsecret', name: 'foxtrot' } },
              }],
            },
          ],
        },
      },
    },
  },
  ingress:: {
    apiVersion: 'networking.k8s.io/v1',
    kind: 'Ingress',
    metadata: {
      namespace: 'foxtrot',
      name: 'foxtrot' + $.config.nameSuffix,
      labels: {
        app: 'foxtrot',
        dev: $.config.dev,
      },
      annotations: {
        'cert-manager.io/cluster-issuer': 'letsencrypt',
        'traefik.ingress.kubernetes.io/router.entrypoints': 'https',
      },
    },
    spec: {
      rules: [
        {
          host: $.config.hostPrefix + $.config.hostname,
          http: {
            paths: [
              {
                backend: {
                  service: {
                    name: 'foxtrot' + $.config.nameSuffix,
                    port: {
                      name: 'http',
                    },
                  },
                },
                path: '/',
                pathType: 'Prefix',
              },
            ],
          },
        },
      ],
      tls: [
        {
          hosts: [$.config.hostPrefix + $.config.hostname],
          secretName: 'foxtrot' + $.config.nameSuffix + '-https-cert',
        },
      ],
    },
  },
}
