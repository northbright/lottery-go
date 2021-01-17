# Example Server

This server is an example lucky draw server which based on [lottery](https://godoc.org/github.com/northbright/lottery-go/lottery) package.

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
* Settings
  * Server settings(`./settings/config.json`)

    The JSON file contains server address and lottery activity name.

    ```
    {
        "addr":":8080",
        "lottery_name":"New Year's Party Lottery"
    }
    ```

  * Participants(`./settings/participants.csv`)

    The CSV file contains records of participants include ID and Name.

    | ID | Name |
    | :--: | :--: |
    | 5 | Fal |
    | 7 | Nango |
    | 8 | Jacky |
    | 9 | Sonny |
    | 10 | Luke |
    | 11 | Mic |
    | 12 | Ric |
    | 13 | Capt |
    | 14 | Andy |
    | 17 | Alex |
    | 33 | Xiao |

  * Prizes(`./settings/prizes.csv`)

    The CSV file is used to set prize's No., name, amount and description.

    | No | Name | Amount | Desc |
    | :--: | :--: | :--: | :--: |
    | 5 | 5th prize | 10 | USB Hard drive |
    | 4 | 4th prize | 8 | Bluetooth Speaker |
    | 3 | 3th prize | 5 | Vacuum Cleaner |
    | 2 | 2nd prize | 2 | Macbook Pro |
    | 1 | 1st prize | 1 | iPhone |

* Run
  
  ```
  // Default server address: :8080
  ./server
  ```

* Test
  * Open browser to vist `http://localhost:8080`
