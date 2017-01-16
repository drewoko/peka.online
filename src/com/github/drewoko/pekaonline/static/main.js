function drawChart(ggpdata){
    var ctx = document.getElementById("myChart");
    var parent = document.getElementById('parent');

    ctx.width = parent.offsetWidth;
    ctx.height = parent.offsetHeight;

    var labels = [];

    var dataset = [{
        label: "Уникальных пользователей на GG",
        data: [],
        borderColor: "#4a7ed0",
        backgroundColor: "rgba(0,0,0,0)"
    }, {
        label: "Количество сообщений на GG",
        data: [],
        borderColor: "#2baed7",
        backgroundColor: "rgba(0,0,0,0)"
    },{
        label: "Уникальных пользователей на Peka",
        data: [],
        borderColor: "#262626",
        backgroundColor: "rgba(0,0,0,0)"
    }, {
        label: "Количество сообщений на Peka",
        data: [],
        borderColor: "#a6a6a6",
        backgroundColor: "rgba(0,0,0,0)"
    }]


    for (i in ggpdata.Goodgame){
        val = ggpdata.Goodgame[i]
        labels.push(moment.unix(val.Timing).format('HH:00'));
        dataset[0].data.push(val.SumUniqUsers)
        dataset[1].data.push(val.SumMessageCount)
    };

    labels[labels.length - 1] = 'latest'

    for (i in ggpdata.Peka){
        val = ggpdata.Peka[i]
        dataset[2].data.push(val.SumUniqUsers)
        dataset[3].data.push(val.SumMessageCount)
    };

    var myChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: dataset
        }
    });
}


function getFormattedDataString(stamp) {
    return moment.unix(stamp * 1).format('DD-MM-YYYY HH:mm');
}

$(function () {
    $.ajax({
        url: "/stats/aggregate"
    }).done(drawChart);

    $.ajax({
        url: "/stats"
    }).done(function(data) {
        var tbody = document.createElement('tbody')
        
        $.each(data, function (i, val) {
            var tr = document.createElement('tr');

            var td = document.createElement('td');
            td.innerHTML = getFormattedDataString(val.Time);
            tr.appendChild(td);

            td = document.createElement('td');
            td.innerHTML = val.GGMessages;
            tr.appendChild(td);

            td = document.createElement('td');
            td.innerHTML = val.PekaMessages;
            tr.appendChild(td);

            td = document.createElement('td');
            td.innerHTML = val.GGUniqUsers;
            tr.appendChild(td);

            td = document.createElement('td');
            td.innerHTML = val.PekaUniqUsers;
            tr.appendChild(td);

            tbody.appendChild(tr);
        });

        $('table').append(tbody);
    });
})