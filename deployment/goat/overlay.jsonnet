{
  manifest+: [$.sealedSecret],
  config+: {
    hostname: 'foxtrot.jul.run',
  },
  sealedSecret:: import 'secret.json',
}
