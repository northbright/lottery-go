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
          v-for="(winner, index) in winners"
          :label="winner.name"
          :key="index"
          :size="fontSize"
          @click="selectWinner(index)"
          :color="red"
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
      currentPrizeIndex: 0,
      currentPrizeContent: "",
      prizes: [],
      availableParticipants: [],
      winners: [],
      started: false,
      drawing: false,
      fontSize: "35px",
      timer: {},
    };
  },

  created() {},

  mounted() {
    var _this = this;
    document.onkeydown = function(e) {
      const key = e.keyCode;
      if (key === 13) {
        window.event.preventDefault();
        _this.startStop();
      }
    };

    this.getPrizes();
  },

  methods: {
    getPrizes() {
      axios
        .get("/prizes")
        .then((response) => {
          this.prizes = response.data.prizes;
          console.log(response);
        })
        .catch((error) => {
          var errMsg = "getPrizes() error: " + error;
          this.$q.notify({
            color: "negative",
            position: "top",
            message: errMsg,
            icon: "report_problem",
          });
          console.log(errMsg);
        });
    },

    getAvailableParticipants(prizeNo) {
      axios
        .post("/available_participants", {
          prize_no: prizeNo,
        })
        .then((response) => {
          console.log(response);
          this.availableParticipants = response.data.available_participants;
          console.log(this.availableParticipants);
        })
        .catch(function(error) {
          var errMsg = "getAvailableParticipants() error: " + error;
          this.$q.notify({
            color: "negative",
            position: "top",
            message: errMsg,
            icon: "report_problem",
          });
          console.log(errMsg);
        });
    },

    getWinners(index) {
      axios
        .post("/winners", {
          prize_no: this.prizes[index].no,
        })
        .then((response) => {
          console.log(response);
          if (response.data.winners.length === 0) {
            var size = this.prizes[index].amount;
            this.winners = [];
            for (var i = 0; i < size; i++) {
              this.winners.push({ id: "?", name: "?" });
            }
          } else {
            this.winners = response.data.winners;
          }
          console.log(this.winners);
        })
        .catch(function(error) {
          var errMsg = "getPrizes() error: " + error;
          this.$q.notify({
            color: "negative",
            position: "top",
            message: errMsg,
            icon: "report_problem",
          });
          console.log(errMsg);
        });
    },

    draw(prizeNo) {
      axios
        .post("/draw", {
          prize_no: prizeNo,
        })
        .then((response) => {
          console.log(response);
          this.winners = response.data.winners;
          console.log(this.winners);
        })
        .catch(function(error) {
          var errMsg = "draw() error: " + error;
          this.$q.notify({
            color: "negative",
            position: "top",
            message: errMsg,
            icon: "report_problem",
          });
          console.log(errMsg);
        });
    },

    selectPrize(index) {
      // Update current prize index and content.
      this.currentPrizeIndex = index;
      this.currentPrizeContent =
        this.prizes[index].name +
        " -- " +
        this.prizes[index].desc +
        "(" +
        this.prizes[index].amount +
        " 人)";

      // Update font accoding to the amount of the prize.
      var prizeNum = this.prizes[index].amount;

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

      // get available participants.
      this.getAvailableParticipants(this.prizes[index].no);

      // get winners.
      this.getWinners(index);
    },

    isCurrentWinnerSelected(index) {
      var idx = this.oldWinnerIndexes.indexOf(index);
      return idx !== -1;
    },

    selectWinner(index) {},

    getButtonColor(index) {
      return this.isCurrentWinnerSelected(index) ? "purple" : "red";
    },

    prizeHasWinners() {
      return this.winners.length === undefined || this.winners.length === 0
        ? false
        : true;
    },

    getPrizeItemClass(prizeIndex) {
      if (prizeIndex === this.currentPrizeIndex) {
        return "bg-red";
      } else {
        return this.prizeHasWinners() ? "bg-pink-2" : "bg-gray";
      }
    },

    startStop() {
      if (!this.started) {
        // Generate random winners
        this.timer = setTimeout(() => {
          for (var i = 0; i < this.prizes[this.currentPrizeIndex].amount; i++) {
            this.winners[i].id = i;
            this.winners[i].name = i;
          }
        }, 10);
        this.started = true;
      } else {
        if (this.drawing) {
          return;
        }

        clearTimeout(this.timer);

        axios
          .post("/draw", {
            prize_no: this.prizes[this.currentPrizeIndex].no,
          })
          .then((response) => {
            console.log(response);

            if (response.data.success) {
              this.winners = response.data.winners;
            } else {
              this.winners = [];
            }

            console.log(this.winners);
            this.drawing = false;
            this.started = false;
          })
          .catch(function(error) {
            var errMsg = "draw error: " + error;
            this.$q.notify({
              color: "negative",
              position: "top",
              message: errMsg,
              icon: "report_problem",
            });
            console.log(errMsg);
            this.drawing = false;
            this.started = false;
          });
        this.drawing = true;
      }
    },
  },
};
</script>
