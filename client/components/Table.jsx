import React, { useState, useEffect } from 'react';
import axios from 'axios';

const TrainTable = () => {
	const [trainData, setTrainData] = useState([]);

	useEffect(() => {
		// Fetch train data from a URL
		axios.get('http://localhost:8080/trains')
			.then(response => setTrainData(response.data))
			.catch(error => console.error('Error fetching data:', error));
	}, []);

	return (
		<div className="container mx-auto mt-8">
		  <h1 className="text-2xl font-bold text-center mb-4">Train Schedule</h1>
		  <table className="w-full border">
			 <thead>
				<tr className="bg-gray-100">
				  <th className="border p-2">Train Name</th>
				  <th className="border p-2">Train Number</th>
				  <th className="border p-2">Departure Time</th>
				  <th className="border p-2">Sleeper Seats</th>
				  <th className="border p-2">AC Seats</th>
				  <th className="border p-2">Sleeper Price</th>
				  <th className="border p-2">AC Price</th>
				  <th className="border p-2">Delayed By</th>
				</tr>
			 </thead>
			 <tbody>
				{trainData.map((train, index) => (
				  <tr key={index} className={(index % 2 === 0) ? 'bg-gray-50' : ''}>
					 <td className="border p-2 text-center justify-center">{train.trainName}</td>
					 <td className="border p-2 text-center ">{train.trainNumber}</td>
					 <td className="border p-2 text-center">
						{train.departureTime.Hours}:{train.departureTime.Minutes}
					 </td>
					 <td className="border p-2 text-center">
						{train.seatsAvailable.sleeper}
					 </td>
					 <td className="border p-2 text-center">
						{train.seatsAvailable.AC}
					 </td>
					 <td className="border p-2 text-center">
						{train.price.sleeper}
					 </td>
					 <td className="border p-2 text-center">
						{train.price.AC}
					 </td>
					 <td className="border p-2 text-center">{train.delayedBy} mins</td>
				  </tr>
				))}
			 </tbody>
		  </table>
		</div>
	 );
};

export default TrainTable;
