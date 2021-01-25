module.exports = {
  printWidth: 80,
  semi: false,
  svelteSortOrder: "scripts-markup-styles",
  tabWidth: 2,
  trailingComma: "all",
  useTabs: false,
  overrides: [
    {
      files: "*.svelte",
      options: {
        parser: "svelte",
      },
    },
  ],
}
