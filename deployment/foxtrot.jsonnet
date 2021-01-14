{
  config:: {
    hostname: null,
    docker_tag: 'latest',
  },
  configure(overlay={}, hostname=null, docker_tag=null)::
    self + overlay + {
      config+: std.prune({
        hostname: hostname,
        docker_tag: docker_tag,
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
      name: 'foxtrot',
      namespace: 'foxtrot',
    },
    spec: {
      ports: [{ name: 'http', port: 8080 }],
      selector: {
        app: 'foxtrot',
      },
    },
  },
  deployment:: {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: {
      labels: {
        app: 'foxtrot',
      },
      name: 'foxtrot',
      namespace: 'foxtrot',
    },
    spec: {
      selector: {
        matchLabels: {
          app: 'foxtrot',
        },
      },
      template: {
        metadata: {
          labels: {
            app: 'foxtrot',
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
      annotations: {
        'cert-manager.io/cluster-issuer': 'letsencrypt',
        'traefik.ingress.kubernetes.io/router.entrypoints': 'https',
      },
      name: 'foxtrot',
      namespace: 'foxtrot',
    },
    spec: {
      rules: [
        {
          host: $.config.hostname,
          http: {
            paths: [
              {
                backend: {
                  service: {
                    name: 'foxtrot',
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
          hosts: [$.config.hostname],
          secretName: 'foxtrot-https-cert',
        },
      ],
    },
  },
}
