/* vim: set ft=javascript sw=2 ts=2 : */

var Gofr = {};

Gofr.helpers = {
	complex: function(r, i) {
		if(typeof r === 'function') { r = r(); }
		if(typeof i === 'function') { i = i(); }
		return r + (i < 0 ? "" : "+") + i + "i";
	}
};

Gofr.FractalBrowser = can.Control.extend({
	defaults: {
		view: '/tmpl/fractal_browser.stache',
		width: 960,
		height: 960,
		bookmarks: {
			home: {
				i: 300,
				e: 4.0,
				m: '#444444',
				c: 'stripe',
				rmin: -2.1,
				rmax: 0.6,
				imin: -1.25,
				imax: 1.25
			},
		},
	}
}, {
	init: function(element, options) {
		var self;

		self = this;

		// TODO use local storage to keep the view
		this.view = new can.Map(this.options.bookmarks.home);
		this.view.attr({
			w: this.options.width,
			h: this.options.height,
		});

		this.view.bind('i', this.update_view.bind(this));
		this.view.bind('e', this.update_view.bind(this));
		this.view.bind('m', this.update_view.bind(this));
		this.view.bind('c', this.update_view.bind(this));
	},
	update: function() {
		var self;

		self = this;

		this.element.html(can.view(
			this.options.view,
			{
				go_home: function(context, element, event) { 
					self.set_view(self.options.bookmarks.home);
					self.update_view();
				},
				move_up: function(context, element, event) { 
					self.translate_view(0.0, -0.0625 * (self.view.imax - self.view.imin));
					self.update_view();
				},
				move_down: function(context, element, event) { 
					self.translate_view(0.0, 0.0625 * (self.view.imax - self.view.imin));
					self.update_view();
				},
				move_left: function(context, element, event) { 
					self.translate_view(-0.0625 * (self.view.rmax - self.view.rmin), 0.0);
					self.update_view();
				},
				move_right: function(context, element, event) { 
					self.translate_view(0.0625 * (self.view.rmax - self.view.rmin), 0.0);
					self.update_view();
				},
				zoom_in: function(context, element, event) { 
					self.scale_view(0.9);
					self.update_view();
				},
				zoom_in_4x: function(context, element, event) { 
					self.scale_view(0.6);
					self.update_view();
				},
				zoom_out: function(context, element, event) { 
					self.scale_view(1.1);
					self.update_view();
				},
				zoom_out_4x: function(context, element, event) { 
					self.scale_view(1.4);
					self.update_view();
				},
				view: this.view,
				url: this.url,
				update_view: this.update_view,
			},
			Gofr.helpers
		));

		this.update_view();
	},
	url: function() {
		return "/png?" + $.param(this.view.serialize());
	},
	update_view: function() {
		var image, self;

		self = this;

		// Find some way to 2-way bind the image URL to the image?
		image = $('#image');
		image.css({height: image.width() + 'px'});
		this.view.attr({
			w: parseInt(image.width()),
			h: parseInt(image.height())
		});

		i = new Image();
		i.onload = function() {
			image.css({
				'background-image': "url(" + self.url() + ")"
			});
		};
		i.src = this.url();
	},
	translate_view: function(r, i) {
		this.view.attr({
			rmin: this.view.rmin += r,
			imin: this.view.imin += i,
			rmax: this.view.rmax += r,
			imax: this.view.imax += i
		});
	},
	scale_view: function(factor) {
		var rw, iw, rmid, imid;

		rw = (this.view.rmax - this.view.rmin) / 2.0;
		iw = (this.view.imax - this.view.imin) / 2.0;
		rmid = this.view.rmin + rw;
		imid = this.view.imin + iw;

		this.view.attr({
			rmin: rmid - (rw * factor),
			imin: imid - (iw * factor),
			rmax: rmid + (rw * factor),
			imax: imid + (iw * factor)
		});
	},
	set_view: function(view) {
		this.view.attr(view);
	},
});

$(document).ready(function() {
	var control;

	control = new Gofr.FractalBrowser('#gofr');
	control.update();
});
