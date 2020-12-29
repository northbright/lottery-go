# Quasar App as Front-End

## Steps
1. Install [Quasar Cli](https://quasar.dev/quasar-cli/installation)
2. Run `quasar create` in `statics` folder
3. Follow the instructions to create Quasar project
   * Input project names
   * Install components(Axios is required)
4. After `quasar create` command done, modify `src/router/routes.js`
   
   Replace `MainLayout.vue` with `MyLayout.vue` for `component`:
   
   `component: () => import('layouts/MyLayout.vue')`

5. Run `quasar build` to build the source code
