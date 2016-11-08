/* vim: set ft=javascript sw=2 ts=2 : */

var Gofr = {};

Gofr.storage = localStorage;

Gofr.helpers = Ractive.defaults.data;
Gofr.helpers.complex = function(r, i) {
		return r + (i < 0 ? "" : "+") + i + "i";
};

Gofr.ModalEditor = Ractive.extend({
	template: '#editor-tmpl',
	data: function() {
		return { 
			key: '',
			text: '',
			mode: 'text',
			theme: 'github',
		};
	},
	oncomplete: function() {
		this.dialog = this.find('dialog');
		this.editor = ace.edit(this.find('div.editor'));
		this.session = this.editor.getSession();
		this.doc = this.session.getDocument();

		this.editor.setTheme("ace/theme/" + this.get('theme'));
		this.session.setMode("ace/mode/" + this.get('mode'));

		// Is this necessary?
		this.editor.$blockScrolling = Infinity;

		this.on({
			save: function() {
				this.set('text', this.doc.getValue());
				this.fire('saved', this.get('key'), this.get('text'));
				this.dialog.close();
			},
			revert: function() {
				this.doc.setValue(this.get('text'));
			},
			dismiss: function() {
				this.dialog.close();
			},
			edit: function(key, text) {
				this.set('key', key);
				this.set('text', text);
				this.doc.setValue(text);
				this.dialog.showModal();
			}
		});
	}
});
Ractive.components.editor = Gofr.ModalEditor;

Gofr.FractalBrowser = Ractive.extend({
	template: '#browser-tmpl',
	data: function() {
		return {
			width: 960,
			height: 960,
			view: {},
			default_view: {
				editable: false,
				i: 300,
				e: 4.0,
				m: '#444444',
				c: 'stripe',
				s: 2.0,
				rmin: -2.1,
				rmax: 0.6,
				imin: -1.25,
				imax: 1.25
			},
			bookmarks: {},
			view_url: this.view_url
		};
	},
	components: {
			editor: Gofr.ModalEditor
	},
	onrender: function() {
			var marks, view;

			view = JSON.parse(Gofr.storage.getItem('gofr.browser.view'));
			if(view) {
				this.set('view', view);
			} else {
				this.copy_view('default_view', 'view');
			}

			marks = JSON.parse(Gofr.storage.getItem('gofr.browser.marks'));
			if(marks) {
				$.each(marks, function(key, value) {
					this.set('bookmarks.' + key, value); 
				}.bind(this));
			} else {
				this.copy_view('default_view', 'bookmarks.home');
			}

			this.socket = new WebSocket('ws://127.0.0.1:8000/png-socket');
			this.socket.onmessage = function(event) {
				console.log('socket: ' + event.data);
			}
			//this.socket.send('wat');
	},
	onunrender: function() {
		if(this.socket) {
			this.socet.close();
		}
	},
	oncomplete: function() {
		this.image = $('#image');

		this.observe('view', function() {
			this.update_view();
			Gofr.storage.setItem('gofr.browser.view', this.json('view'));
		});

		this.observe('bookmarks', function() {
			Gofr.storage.setItem('gofr.browser.marks', this.json('bookmarks'));
		});

		this.on({
			move_up: function() {
				this.translate_view(0.0, -0.0625 * (this.get('view.imax') - this.get('view.imin')));
			},
			move_down: function() {
				this.translate_view(0.0, 0.0625 * (this.get('view.imax') - this.get('view.imin')));
			},
			move_left: function() {
				this.translate_view(-0.0625 * (this.get('view.rmax') - this.get('view.rmin')), 0.0);
			},
			move_right: function() {
				this.translate_view(0.0625 * (this.get('view.rmax') - this.get('view.rmin')), 0.0);
			},
			zoom_in: function() {
				this.scale_view(0.9);
			},
			zoom_in_4x: function() {
				this.scale_view(0.6);
			},
			zoom_out: function() {
				this.scale_view(1.1);
			},
			zoom_out_4x: function() {
				this.scale_view(1.4);
			},	
			update_view: function() {
				this.update_view();
			},
			go_to_bookmark: function(event) {
				var bookmark, name;
				 
				name = 'bookmarks.' + $(event.node).data('bookmark');
				if(!name in this.get('bookmarks')) { return; }
				this.copy_view(name, 'view');
			},
			add_bookmark: function(event) {
				var name;

				// TODO FIX THIS window.prompt 
				name = window.prompt('Enter a name:');

				if(name) {
					this.copy_view('view', 'bookmarks.' + name);
					this.set('bookmarks.' + name + '.editable', true); 
				}
			},
			update_bookmark: function(event) {
				var name;

				name = $(event.node).data('bookmark');
				this.copy_view('view', 'bookmarks.' + name);
				this.set('bookmarks.' + name + '.editable', true); 
			},
			delete_bookmark: function(event) {
				var bookmarks;

				bookmarks = this.get('bookmarks');
				delete bookmarks[$(event.node).data('bookmark')];
				this.update('bookmarks');
			},
			edit_view: function() {
				var editor;

				editor = this.findComponent('editor');
				editor.fire('edit', 'view', this.json('view'));
			},
			edit_bookmarks: function() {
				var editor;

				editor = this.findComponent('editor');
				editor.fire('edit', 'bookmarks', this.json('bookmarks'));
			},
			'editor.saved': function(key, text) {
				this.set(key, JSON.parse(text));
			}
		});
	},
	json: function(key) {
		return JSON.stringify(this.get(key), null, 2);
	},
	view_url: function(name) {
		return "/png?" + $.param(this.get('view'));
	},
	update_size: function() {
		this.image.css({height: this.image.width() + 'px'});
		this.set('view.w', parseInt(this.image.width()));
		this.set('view.h', parseInt(this.image.height()));
	},
	update_view: function() {
		var i;

		this.update_size();
		i = new Image();
		i.onload = function() {
			this.image.css({
				'background-image': "url(" + this.view_url() + ")"
			});
		}.bind(this);

		i.src = this.view_url();
	},
	translate_view: function(r, i) {
		var view;

		view = this.get('view');

		view.rmin += r;
		view.imin += i;
		view.rmax += r;
		view.imax += i;

		this.update('view');
	},
	scale_view: function(factor) {
		var rw, iw, rmid, imid, view;

		view = this.get('view');

		rw = (view.rmax - view.rmin) / 2.0;
		iw = (view.imax - view.imin) / 2.0;
		rmid = view.rmin + rw;
		imid = view.imin + iw;

		view.rmin = rmid - (rw * factor);
		view.imin = imid - (iw * factor);
		view.rmax = rmid + (rw * factor);
		view.imax = imid + (iw * factor);

		this.update('view');
	},
	copy_view: function(src_key, dst_key) {
		var view;

		view = this.get(src_key);

		this.set(dst_key, {
			w: view.w,
			h: view.h,
			i: view.i,
			e: view.e,
			m: view.m,
			c: view.c,
			s: view.s,
			rmin: view.rmin,
			rmax: view.rmax,
			imin: view.imin,
			imax: view.imax,
		});
	},
});
Ractive.components.browser = Gofr.FractalBrowser;

$(document).ready(function() {
	var ractive = Ractive({
		el: '#gofr',
		template: '#gofr-tmpl',
		data: function() { return {}; },
		components: {
			browser: Gofr.FractalBrowser,
		},
	});
});

