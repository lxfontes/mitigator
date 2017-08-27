var vm = new Vue({
  el: '#vue-instance',
  data: {
    txt: "The designer's weblog astonishing commitment to youth textures has been a game-changer.\nA webinar will be amazing.",
    hits: null,
    inFlight: false,
    input: '',
  },
  methods: {
    checkWebsite: function() {
      this.inFlight = true;
      this.input = this.txt;
      this.hits = null;
      axios.post("/score", { txt: this.input })
        .then(response => {
          this.hits = response.data.hits;
          console.log(response);
        }).catch(function (error) {
          console.log(error);
        });
    }},
  computed: {
    markedHits: function () {
      if (this.hits == null) {
        return '';
      }

      function escapeRegExp(str) {
        // ohai stackoverflow
        return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"); // $& means the whole matched string
      }

      var wrapHit = function(txt, str, tags) {
        var classes = [];
        for (var h in tags) {
          var tag = tags[h];
          if (tag != "english") {
            var replacement = "<span class='tag is-warning'>" + str + "</span>";
            var search = '\\b' + escapeRegExp(str);
            return txt.replace(new RegExp(search, 'g'), replacement);
          }
        }

        // only tag non-english
        return txt;
      };

      var markedText = this.input;
      for (var h in this.hits) {
        var hit = this.hits[h];
        markedText = wrapHit(markedText, h, hit)
      }

      return markedText.replace(/$/mg,'<br>');;
    }
  }
});
