{
  manifest+: [if self.config.dev == '' then $.sealedSecret],
  config+: {
    hostname: 'foxtrot.jul.run',
  },
  sealedSecret:: import 'secret.json',
}
