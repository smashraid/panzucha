const MenuApp = {
    delimiters: ['${', '}'],
    data() {
      return {
        counter: 0,
        count: 0,
        items: []
      }
    },
    created() {
      // Debouncing with Lodash
      //this.debouncedClick = _.debounce(this.click, 500)
    },    
    mounted() {
      // setInterval(() => {
      //   this.counter++
      // }, 1000)
    },
    unmounted() {
      // Cancel the timer when the component is removed
      //this.debouncedClick.cancel()
    },
    methods: {
      say(message) {
        alert(message)
      },
      increment() {
        // `this` will refer to the component instance
        this.count++
      },
      add(product) {
        let itemIndex = this.items.findIndex(item => item.id == product.id);
        console.log('index', itemIndex);
        if (itemIndex > 0) {
          this.items[itemIndex].quantity++;
        } else {
          this.items.push(product);
        }        
      }
    }
  }

  Vue.createApp(MenuApp).mount('#menu')