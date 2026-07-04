(function () {
  var mapInstance = null;

  function readMarkers() {
    var el = document.getElementById("mercado-map-data");
    if (!el || !el.textContent) return [];
    try {
      var data = JSON.parse(el.textContent);
      return Array.isArray(data) ? data : [];
    } catch (e) {
      return [];
    }
  }

  function defaultCenter(markers) {
    if (!markers.length) {
      return { lat: -23.542, lng: -46.565, zoom: 13 };
    }
    var lat = 0;
    var lng = 0;
    markers.forEach(function (m) {
      lat += m.lat;
      lng += m.lng;
    });
    return { lat: lat / markers.length, lng: lng / markers.length, zoom: 13 };
  }

  function popupHTML(m) {
    var deal = m.bestDeal ? "<br><strong>" + m.bestDeal + "</strong>" : "";
    return (
      "<div class='mercado-map-popup'>" +
      "<strong>" +
      m.name +
      "</strong><br>" +
      "<span>" +
      m.address +
      "</span><br>" +
      "<span>" +
      m.offers +
      " ofertas" +
      deal +
      "</span></div>"
    );
  }

  function fixLeafletIcons() {
    if (!window.L || !L.Icon || !L.Icon.Default) return;
    L.Icon.Default.mergeOptions({
      iconUrl: "/static/img/leaflet/marker-icon.png",
      iconRetinaUrl: "/static/img/leaflet/marker-icon-2x.png",
      shadowUrl: "/static/img/leaflet/marker-shadow.png",
    });
  }

  function destroyMap() {
    if (mapInstance) {
      mapInstance.remove();
      mapInstance = null;
    }
  }

  function initMap() {
    var container = document.getElementById("mercado-map");
    if (!container || !window.L) return;

    destroyMap();

    var markers = readMarkers();
    var center = defaultCenter(markers);
    fixLeafletIcons();

    mapInstance = L.map(container, { scrollWheelZoom: false }).setView(
      [center.lat, center.lng],
      center.zoom,
    );

    L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    }).addTo(mapInstance);

    var bounds = [];
    markers.forEach(function (m) {
      var marker = L.marker([m.lat, m.lng]).addTo(mapInstance);
      marker.bindPopup(popupHTML(m));
      bounds.push([m.lat, m.lng]);
    });

    if (bounds.length > 1) {
      mapInstance.fitBounds(bounds, { padding: [24, 24], maxZoom: 14 });
    }

    setTimeout(function () {
      if (mapInstance) mapInstance.invalidateSize();
    }, 100);
  }

  window.MercadoMap = {
    init: initMap,
    destroy: destroyMap,
  };

  document.addEventListener("DOMContentLoaded", initMap);
})();