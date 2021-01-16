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
          style="width: 200px"
          size="20px"
          :label="getDrawButtonLabel()"
          @click="onDrawButtonClick()"
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
          :color="getWinnerColor(index)"
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
      // Record selected winner indexes to change these winners with fake winners when redraw.
      selectedWinnerIndexes: [],
      // Winners be overwritten while generating fake winners,
      // recording selected winner indexes is not enough,
      // need also make a copy of selected winners.
      selectedWinners: [],
      started: false,
      drawing: false,
      fontSize: "35px",
      timer: {},
    };
  },

  created() {},

  mounted() {
    var _this = this;
    document.onkeyup = function(e) {
      const key = e.keyCode;
      if (key === 13) {
        window.event.preventDefault();
        _this.onDrawButtonClick();
      }
    };

    this.getPrizes();
  },

  methods: {
    notify(msg) {
      this.$q.notify({
        color: "purple",
        message: msg,
      });
    },

    getRandomInt(min, max) {
      min = Math.ceil(min);
      max = Math.floor(max);
      return Math.floor(Math.random() * (max - min + 1)) + min;
    },

    getRandomFakeWinners(availableParticipants, amount) {
      if (availableParticipants.length <= 0) {
        return [];
      }

      if (amount <= 0) {
        return [];
      }

      var newAmount =
        availableParticipants.length < amount
          ? availableParticipants.length
          : amount;
      var newAvailableParticipants = availableParticipants;

      var fakeWinners = [];

      while (fakeWinners.length < newAmount) {
        var minIndex = 0;
        var maxIndex = newAvailableParticipants.length - 1;

        var index = this.getRandomInt(minIndex, maxIndex);
        var fakeWinner = newAvailableParticipants[index];
        fakeWinners.push(fakeWinner);

        // Remove winner from available participants.
        newAvailableParticipants = newAvailableParticipants.filter(
          (item) => item !== fakeWinner
        );
      }

      return fakeWinners;
    },

    getPrizes() {
      axios
        .get("/prizes")
        .then((response) => {
          if (response.data.success) {
            this.prizes = response.data.prizes;
          } else {
            var errMsg = "/prizes error: " + response.data.err_msg;
            this.notify(errMsg);
          }
        })
        .catch((e) => {
          var errMsg = "/prizes axios error: " + e;
          this.notify(errMsg);
        });
    },

    getAvailableParticipants(prizeNo) {
      axios
        .post("/available_participants", {
          prize_no: prizeNo,
        })
        .then((response) => {
          if (response.data.success) {
            this.availableParticipants = response.data.available_participants;
          } else {
            var errMsg =
              "/available_participants error: " + response.data.err_msg;
            this.notify(errMsg);
          }
        })
        .catch((e) => {
          var errMsg = "/available_participants axios error: " + e;
          this.notify(errMsg);
        });
    },

    getWinners(index) {
      axios
        .post("/winners", {
          prize_no: this.prizes[index].no,
        })
        .then((response) => {
          if (response.data.success) {
            this.winners = response.data.winners;
          } else {
            var errMsg = "/winners error: " + response.data.err_msg;
            this.notify(errMsg);
          }
        })
        .catch((e) => {
          var errMsg = "/winners axios error: " + e;
          this.notify(errMsg);
        })
        .then(() => {
          if (this.winners.length === 0) {
            var size = this.prizes[index].amount;
            this.winners = [];
            for (var i = 0; i < size; i++) {
              this.winners.push({ id: "?", name: "?" });
            }
          }
        });
    },

    draw(prizeNo) {
      if (this.drawing) {
        var msg = "is drawing...please wait";
        this.notify(msg);
        return;
      }

      this.drawing = true;

      axios
        .post("/draw", {
          prize_no: prizeNo,
        })
        .then((response) => {
          if (response.data.success) {
            this.winners = response.data.winners;
          } else {
            var errMsg = "/draw error: " + response.data.err_msg;
            this.notify(errMsg);
          }
        })
        .catch((e) => {
          var errMsg = "/draw axios error: " + e;
          this.notify(errMsg);
        })
        .then(() => {
          this.started = false;
          this.drawing = false;
        });
    },

    revokeAndRedraw(prizeNo) {
      if (this.drawing) {
        var msg = "is re-drawing...please wait";
        this.notify(msg);
        return;
      }

      this.drawing = true;

      console.log("selecteWinners:");
      console.log(this.selectedWinners);

      axios
        .post("/revoke", {
          prize_no: prizeNo,
          revoked_winners: this.selectedWinners,
        })
        .then((response) => {
          if (response.data.success) {
            // Revoke winners successfully, re-draw new winners
            axios
              .post("/redraw", {
                prize_no: prizeNo,
                amount: this.selectedWinners.length,
              })
              .then((response) => {
                var newWinners = response.data.winners;

                // Replace revoked winners with new winners
                for (var i = 0; i < this.selectedWinnerIndexes.length; i++) {
                  this.winners.splice(
                    this.selectedWinnerIndexes[i],
                    1,
                    newWinners[i]
                  );
                }
              })
              .catch((e) => {
                var errMsg = "/redraw axios error: " + e;
                this.notify(errMsg);
              })
              .then(() => {
                this.started = false;
                this.drawing = false;
              });
          } else {
            var errMsg = "/revoke error: " + response.data.err_msg;
            this.notify(errMsg);
          }
        })
        .catch((e) => {
          var errMsg = "/revoke axios error: " + e;
          this.notify(errMsg);
        })
        .then(() => {
          this.started = false;
          this.drawing = false;
        });
    },

    selectPrize(index) {
      // Clear timer if needed.
      if (this.started) {
        clearInterval(this.timer);
        this.started = false;
      }

      // Clear selected winners.
      this.selectedWinnerIndexes = [];
      this.selectedWinners = [];

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
      return this.selectedWinnerIndexes.indexOf(index) !== -1;
    },

    selectWinner(index) {
      if (this.started) {
        var errMsg = "Can not select winner when drawing or redrawing...";
        this.notify(errMsg);
        return;
      }

      if (this.winners[index].id === "?") {
        return;
      }

      if (this.isCurrentWinnerSelected(index)) {
        // Remove winner index from selected winner indexes.
        this.selectedWinnerIndexes = this.selectedWinnerIndexes.filter(
          (item) => item !== index
        );
      } else {
        this.selectedWinnerIndexes.push(index);
      }
    },

    getWinnerColor(index) {
      return this.isCurrentWinnerSelected(index) ? "purple" : "red";
    },

    currentPrizeHasWinners() {
      return this.winners.length === undefined ||
        this.winners.length === 0 ||
        this.winners[0].id === "?"
        ? false
        : true;
    },

    getPrizeItemClass(prizeIndex) {
      if (prizeIndex === this.currentPrizeIndex) {
        return "bg-red";
      } else {
        return "bg-pink-2";
      }
    },

    hasSelectedWinners() {
      return this.selectedWinnerIndexes.length !== 0 ? true : false;
    },

    getDrawButtonLabel() {
      if (this.started !== true) {
        if (!this.hasSelectedWinners()) {
          return "开始";
        } else {
          return "开始重抽";
        }
      } else {
        return "停止";
      }
    },

    onDrawButtonClick() {
      // Draw
      if (!this.hasSelectedWinners()) {
        if (!this.started) {
          if (this.currentPrizeHasWinners()) {
            var errMsg = "Winners of current prize exist";
            this.notify(errMsg);
            return;
          }

          if (this.availableParticipants.length === 0) {
            var errMsg = "No available participants for current prize";
            this.notify(errMsg);
            return;
          }

          // Generate random winners
          this.timer = setInterval(() => {
            var amount = this.prizes[this.currentPrizeIndex].amount;
            this.winners = this.getRandomFakeWinners(
              this.availableParticipants,
              amount
            );
          }, 100);
          this.started = true;
        } else {
          clearInterval(this.timer);

          this.draw(this.prizes[this.currentPrizeIndex].no);
        }
      } else {
        // Revoke and re-draw
        if (!this.started) {
          if (!this.hasSelectedWinners()) {
            var errMsg = "No selected winners to revoke";
            this.notify(errMsg);
            return;
          }

          if (this.availableParticipants.length === 0) {
            var errMsg = "No available participants for current prize";
            this.notify(errMsg);
            return;
          }

          // Copy selected winners.
          this.selectedWinners = [];
          for (var i = 0; i < this.selectedWinnerIndexes.length; i++) {
            this.selectedWinners.push(
              this.winners[this.selectedWinnerIndexes[i]]
            );
          }

          console.log("available participants: ");
          console.log(this.availableParticipants);

          // Generate random winners
          this.timer = setInterval(() => {
            var amount = this.selectedWinnerIndexes.length;
            var fakeRedrawWinners = this.getRandomFakeWinners(
              this.availableParticipants,
              amount
            );

            console.log("fakeRedrawWinners:");
            console.log(fakeRedrawWinners);

            for (var i = 0; i < fakeRedrawWinners.length; i++) {
              this.winners.splice(
                this.selectedWinnerIndexes[i],
                1,
                fakeRedrawWinners[i]
              );

              console.log(
                "this.selectedWinnerIndexes[i]: " +
                  this.selectedWinnerIndexes[i]
              );
              console.log("fakeRedrawWinners[i]: ");
              console.log(fakeRedrawWinners[i]);
            }
          }, 100);
          this.started = true;
        } else {
          clearInterval(this.timer);

          this.revokeAndRedraw(this.prizes[this.currentPrizeIndex].no);
        }
      }
    },
  },
};
</script>
