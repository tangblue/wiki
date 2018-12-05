export default {
    template: `
    <div>
        <button v-on:click="count++">You clicked me {{ count }} times.</button>
        </br>
        <button v-on:click="$emit('enlarge-text', 0.1)">Enlarge text</button>
        <button v-on:click="showCount">Show count</button>
    </div>`,
    data: function () {
        return {
            count: 0
        }
    },
    methods: {
        showCount: function() {
            console.log(this.count);
        }
    }
}
