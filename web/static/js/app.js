(function ($) {
  'use strict';

  var menuItems = [];
  var idToItem = {};

  function indexMenu(items, parent) {
    items.forEach(function (item) {
      idToItem[item.id] = { item: item, parent: parent };
      if (item.children && item.children.length) {
        indexMenu(item.children, item.id);
      }
    });
  }

  function firstLeaf(items) {
    for (var i = 0; i < items.length; i++) {
      var it = items[i];
      if (it.children && it.children.length) {
        var nested = firstLeaf(it.children);
        if (nested) return nested;
      } else if (it.page) {
        return it;
      }
    }
    return null;
  }

  function renderMenu(items, $container, depth) {
    items.forEach(function (item) {
      if (item.children && item.children.length) {
        $container.append(
          $('<div class="menu-group-toggle"></div>').text(item.title)
        );
        var $sub = $('<div class="submenu nav flex-column"></div>');
        renderMenu(item.children, $sub, depth + 1);
        $container.append($sub);
        return;
      }
      var icon = item.icon ? '<i class="bi ' + item.icon + ' me-2"></i>' : '';
      var $link = $('<a class="nav-link" href="#"></a>')
        .attr('data-id', item.id)
        .attr('data-page', item.page)
        .html(icon + $('<span></span>').text(item.title).html());
      $container.append($link);
    });
  }

  function setActive(id) {
    $('#sidebar-nav .nav-link').removeClass('active');
    $('#sidebar-nav .nav-link[data-id="' + id + '"]').addClass('active');
    var meta = idToItem[id];
    if (meta) {
      $('#page-title h1').text(meta.item.title);
    }
  }

  function showError(msg) {
    $('#main-content').html(
      '<div class="alert alert-danger" role="alert">' +
        $('<span></span>').text(msg).html() +
        '</div>'
    );
  }

  function loadPage(item) {
    if (!item || !item.page) return;
    setActive(item.id);
    location.hash = '#/' + item.id;

    $('#main-content').html(
      '<div class="text-center text-muted py-5"><div class="spinner-border" role="status"></div></div>'
    );

    $('#main-content').load('/pages/' + item.page, function (response, status) {
      if (status === 'error') {
        showError('无法加载页面: ' + item.page);
      }
    });
  }

  function routeFromHash() {
    var hash = location.hash.replace(/^#\/?/, '');
    if (hash && idToItem[hash] && idToItem[hash].item.page) {
      loadPage(idToItem[hash].item);
      return;
    }
    var leaf = firstLeaf(menuItems);
    if (leaf) loadPage(leaf);
  }

  function bindNav() {
    $('#sidebar-nav').on('click', '.nav-link[data-page]', function (e) {
      e.preventDefault();
      var id = $(this).attr('data-id');
      if (idToItem[id]) loadPage(idToItem[id].item);
    });
    $(window).on('hashchange', routeFromHash);
  }

  function init() {
    $.getJSON('/api/menu')
      .done(function (data) {
        menuItems = data.items || [];
        idToItem = {};
        indexMenu(menuItems, null);
        var $nav = $('#sidebar-nav').empty();
        renderMenu(menuItems, $nav, 0);
        bindNav();
        routeFromHash();
      })
      .fail(function () {
        showError('无法加载菜单配置，请检查服务是否正常运行。');
      });
  }

  $(init);
})(jQuery);
