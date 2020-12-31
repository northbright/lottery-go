# Example Server

## Front-End
* [Quasar](https://quasar.dev/) SPA
* How to build
  1. Install [Quasar Cli](https://quasar.dev/quasar-cli/installation)
  2. Run `quasar create` in `statics` folder
     
     ```
     cd ./statics
     quasar create
     ```

  3. Follow the instructions to create Quasar project
     * Input project names
     * Install Axios
     * Select Prettier as the linter for ESLint
  4. After `quasar create` command done, modify `src/router/routes.js`

     Replace `MainLayout.vue` with `MyLayout.vue` for `component`:

     `component: () => import('layouts/MyLayout.vue')`

  5. [Install Notify Quasar Plugin](https://quasar.dev/quasar-plugins/notify#Installation)

     ```
     // quasar.conf.js

     return {
       framework: {
         plugins: [
           'Notify'
         ],
         config: {
           notify: { /* look at QUASARCONFOPTIONS from the API card (bottom of page) */ }
         }
       }
     }
     ```

  6. Run `quasar build` to build the source code
     * It will put the release under `/dist/spa`

## Back-End
* Go HTTP server which provide lottery service
* How to build
  
  ```
  go build
  ```
* Run
  
  ```
  // Default server address: :8080
  ./server
  ```

* Test
  * Open browser to vist `http://localhost:8080`
