var state = null;

function list_item_click( event ) {
	let task_id = event.target.getAttribute("task-id");
	let clicked_task = state.tasks.pending.find( task => task.id == task_id );
	
	console.log( task_id,  clicked_task );
	if( clicked_task ) {
		state.task_id_clicked = task_id;
		open_modal( clicked_task );
	}

}

function mark_task_done( task_id, return_to_pool ) {
	
	
	let task = state.tasks.pending.find( task => task.id == task_id );
	
	// remove this instead of setting to true since we don't care about it being done as such
	delete task["done"];
	
	state.tasks.pending = state.tasks.pending.filter( task => task.id != task_id );
	console.log( task_id, return_to_pool, task );
	
	// either return it to the pool so it can come up again, or stash it away
	// so you won't see it again this (week|month)?
	if( return_to_pool ) {
		state.tasks.available.push( task );
	} else {
		state.tasks.stashed.push( task );
	}
	
	save( state );
	console.log("markdone", state);

	render( state );

	close_modal();
}

function open_modal( task ) {
	
	let modal_title = document.getElementById("modal-task-title");
	modal_title.innerHTML = task.name;

	["modal", "modal-overlay"].forEach( dom_id => document.getElementById(dom_id).classList.toggle("closed") );
}

function close_modal() {

	["modal", "modal-overlay"].forEach( dom_id => document.getElementById(dom_id).classList.toggle("closed") );
}

function setup_modal() {
	
	let close_button = document.querySelector("#close-button");
	let done_return_button = document.querySelector("#done-return-button");
	let done_stash_button = document.querySelector("#done-stash-button");

	close_button.onclick = close_modal;
	done_return_button.onclick = function() { mark_task_done( state.task_id_clicked, true ) };
	done_stash_button.onclick = function() { mark_task_done( state.task_id_clicked, false ) };
}

function clear( element ) {
	
	while( element.hasChildNodes() ) {
		element.removeChild( element.firstChild );
	}
}

function make_list( dom_id, tasks ) {
	
	let list = document.getElementById( dom_id );
	clear( list );
	
	let list_items = tasks.map( task => {
		let li = document.createElement("li");
		li.setAttribute("task-id", task.id );
		li.onclick = list_item_click;
		li.appendChild( document.createTextNode( task.name ) );
		return li;
	});
	
	list_items.forEach( li => list.appendChild( li) );
}

function render( state ) {
	
	let d = new Date( state.current_day );
	var options = { weekday: 'long', month: 'long', day: 'numeric' };
	console.log("DATEFORMAT", d );
	console.log(d.toLocaleString('en-US', options));
	document.getElementById("current_day").innerHTML = d.toLocaleString('en-US', options);
	
	make_list( 'available', state.tasks.available );
	make_list( 'pending', state.tasks.pending );
	make_list( 'stashed', state.tasks.stashed );

	let overdue = state.tasks.pending.filter( task => !task.weekly && task.day < state.current_day );
	let overdue_container = document.getElementById("overdue");
	if( overdue.length == 0 ) {
		overdue_container.classList.add("closed");
	} else {
		overdue_container.classList.remove("closed");
		make_list( 'overdue_items', overdue );
	}
	
	let today = state.tasks.pending.filter( task => task.day == state.current_day );
	console.log( "Render", today );
	if( today.length == 0 ) {
		if( state.tasks.pending.length > 0 ) {
			today.push( { "name": state.tasks.pending.length + " pending!" } );			
		} else {
			today.push( { "name": "Nothing to do!" } );			
		}
	}
	make_list( 'today', today );
	
	let week = state.tasks.pending.filter( task => task.weekly );
	let week_container = document.getElementById("this_week");
	if( week.length == 0 ) {
		week_container.classList.add("closed");
	} else {
		week_container.classList.remove("closed");
		make_list( 'week_items', week );
	}
	

	document.getElementById("state").innerHTML = JSON.stringify( state, null, "\t" );
}



function go() {

}
