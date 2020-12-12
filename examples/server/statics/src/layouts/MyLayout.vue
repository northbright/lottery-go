<template>
  <q-layout view="lHh Lpr lff">
    <q-header elevated>
      <q-toolbar class="bg-pink-3 text-black">
        <q-btn
          flat
          dense
          round
          @click="leftDrawerOpen = !leftDrawerOpen"
          icon="menu"
          aria-label="Menu"
        />

        <q-toolbar-title>{{ currentPrizeContent }} </q-toolbar-title>

        <div>Quasar v{{ $q.version }}</div>
      </q-toolbar>
    </q-header>

    <q-footer bordered class="bg-white text-primary">
      <div class="q-pa-md q-gutter-xl fit row wrap justify-around">
        <q-btn
          rounded
          color="pink-3"
          text-color="black"
          type="submit"
          size="20px"
          label="start/stop"
          @click="startStop()"
        />
      </div>
    </q-footer>

    <q-drawer
      v-model="leftDrawerOpen"
      show-if-above
      bordered
      content-class="bg-grey-2"
    >
      <q-list>
        <q-item-label header>奖项</q-item-label>
        <q-item
          v-for="(prize, index) in prizes"
          :key="index"
          clickable
          @click="selectPrize(index)"
          :class="getPrizeItemClass(index)"
        >
          <q-item-section avatar>
            <q-icon name="school" />
          </q-item-section>
          <q-item-section>
            <q-item-label>{{ prize.name }}</q-item-label>
          </q-item-section>
        </q-item>
      </q-list>
    </q-drawer>

    <q-page-container>
      <div
        class="q-gutter-xl fit row wrap justify-around content-center bg-pink-1"
      >
        <q-btn
          v-for="(winner, index) in this.winners"
          :label="winner.name"
          :key="index"
          :size="fontSize"
          @click="selectWinner(index)"
          :color="getButtonColor(index)"
        ></q-btn>
      </div>
    </q-page-container>
  </q-layout>
</template>

<script>
import axios from "axios";

export default {
  name: "MyLayout",

  data() {
    return {
      leftDrawerOpen: false,
      currentPrizeContent: "",
      prizes: [],
      prizeNo: 0,
      prizesDone: [],
      winners: [],
      oldWinnerIndexes: [],
      url: "",
      conn: {},
      started: false,
      fontSize: "35px"
    };
  },

  created() {
    console.log("created()");
    //this.initWebSocket();
    this.loadData();
  },

  mounted() {
    var _this = this;
    document.onkeydown = function(e) {
      let key = e.keyCode;
      if (key == 13) {
        window.event.preventDefault();
        _this.startStop();
      }
    };
  },

  methods: {
    loadData() {
      axios
        .get("/get-ws-url")
        .then(response => {
          this.url = response.data;
          console.log(response);

          console.log(this.url);
          this.initWebSocket(this.url);
        })
        .catch(() => {
          this.$q.notify({
            color: "negative",
            position: "top",
            message: "Loading failed",
            icon: "report_problem"
          });
        });
    },

    initWebSocket(url) {
      console.log("initWebSocket()");
      this.conn = new WebSocket(url);
      this.conn.onopen = this.webSocketOnOpen;
      this.conn.onmessage = this.webSocketOnMessage;
    },

    webSocketOnOpen() {
      console.log("WebSocket on open");
      var action = { name: "get_prizes" };
      this.conn.send(JSON.stringify(action));
    },

    webSocketOnMessage(msg) {
      var res = JSON.parse(msg.data);
      console.log(res);
      if (res.success === true) {
        var action = res.action;

        switch (action.name) {
          case "get_prizes":
            this.prizes = res.prizes;
            break;

          case "get_winners":
            if (res.winners.length === 0) {
              this.winners = [];
              var prizeNum = this.prizes[this.prizeNo].num;

              for (var i = 0; i < prizeNum; i++) {
                this.winners[i] = { id: "", name: "?" };
              }
            } else {
              this.winners = res.winners;
            }

            break;

          case "start":
            this.started = true;
            this.winners = res.winners;
            break;

          case "stop":
            this.started = false;
            this.winners = res.winners;
            this.prizesDone[res.action.prize_index] = true;
            break;
        }
      }
    },

    selectPrize(index) {
      this.prizeNo = index;
      this.currentPrizeContent =
        this.prizes[this.prizeNo].name +
        " -- " +
        this.prizes[this.prizeNo].content;

      console.log("prize: " + index + "selected");
      var action = { name: "get_winners", prize_index: index };
      this.conn.send(JSON.stringify(action));
      this.oldWinnerIndexes = [];

      var prizeNum = this.prizes[index].num;

      if (prizeNum >= 20) {
        this.fontSize = "35px";
      } else if (prizeNum >= 10) {
        this.fontSize = "50px";
      } else if (prizeNum >= 5) {
        this.fontSize = "60px";
      } else if (prizeNum >= 2) {
        this.fontSize = "70px";
      } else {
        this.fontSize = "80px";
      }
    },

    isCurrentWinnerSelected(index) {
      var idx = this.oldWinnerIndexes.indexOf(index);
      return idx === -1 ? false : true;
    },

    selectWinner(index) {
      console.log(this.oldWinnerIndexes);
      if (this.isCurrentWinnerSelected(index)) {
        var idx = this.oldWinnerIndexes.indexOf(index);
        console.log("idx: " + idx);

        console.log("selected: before remove:");
        console.log(this.oldWinnerIndexes);

        //delete this.oldWinnerIndexes[idx];
        this.oldWinnerIndexes.splice(idx, 1);

        console.log("selected: after remove:");
        console.log(this.oldWinnerIndexes);
      } else {
        console.log("not selected: before push");
        console.log(this.oldWinnerIndexes);

        this.oldWinnerIndexes.push(index);

        console.log("not selected: after push");
        console.log(this.oldWinnerIndexes);
      }
    },

    getButtonColor(index) {
      return this.isCurrentWinnerSelected(index) ? "purple" : "red";
    },

    getPrizeItemClass(index) {
      if (index == this.prizeNo) {
        return "bg-red";
      } else {
        return this.prizesDone[index] ? "bg-pink-2" : "bg-gray";
      }
    },

    startStop() {
      var action = {};

      if (!this.started) {
        action.name = "start";
      } else {
        action.name = "stop";
      }

      this.started = !this.started;

      action.prize_index = this.prizeNo;
      action.old_winner_indexes = this.oldWinnerIndexes;
      console.log(action);
      this.conn.send(JSON.stringify(action));
    }
  }
};
</script>
