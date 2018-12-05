import Counter from './Counter.js';

const User = { template: '<div>User {{ $route.params.id + $route.query.test }}</div>' }

const routes = [
  { path: '/user/:id', component: User },
  { path: '/counter', component: Counter }
]

const router = new VueRouter({
  routes
})

var app = new Vue({
    el: '#app',
    template: `
    <div>
          <h1>Hello App!</h1>
          <input :value="message" @input="update"></input>
          <button v-on:click="showMessage">Show</button>
          </br>
          {{ message }}
          </br>
          <p>
            <router-link :to="{ path: '/user/'+ $route.query.test }">Go to User</router-link>
            <router-link to="/counter">Go to Counter</router-link>
          </p>
          <keep-alive>
            <router-view v-on:enlarge-text="this.console.log($event)"></router-view>
          </keep-alive>
    </div>`,
    router,
    data: {
        message: 'Hello Vue!'
    },
    methods: {
        update: _.debounce(function (e) {
            axios.get('https://yesno.wtf/api')
                .then(function (response) {
                  this.message = e.target.value + _.capitalize(response.data.answer)
                }.bind(this))
                .catch(function (error) {
                  this.message = 'Error! Could not reach the API. ' + error
                }.bind(this))
        }, 300),
        showMessage: _.debounce(function(e) {
            console.log(this.message);
        }, 300)
    }
});
